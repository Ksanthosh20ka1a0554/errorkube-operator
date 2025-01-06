/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	errorkubev1 "github.com/ksanthosh20ka1a0554/errorkube-operator/api/v1"
)

// ErrorKubeReconciler reconciles a ErrorKube object
type ErrorKubeReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=errorkube.errorkube.io,resources=errorkubes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=errorkube.errorkube.io,resources=errorkubes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=errorkube.errorkube.io,resources=errorkubes/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=rbac.authorization.k8s.io,resources=clusterroles;clusterrolebindings,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services;serviceaccounts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups="",resources=events,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=pods/log,verbs=get;list;watch

// Reconcile logic set the current state of the cluster closer to the desired state.
func (r *ErrorKubeReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	// Fetch the ErrorKube instance
	errorkube := &errorkubev1.ErrorKube{}
	if err := r.Get(ctx, req.NamespacedName, errorkube); err != nil {
		log.Error(err, "Unable to fetch ErrorKube instance")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

// Define the Service Account resource
serviceAccount := &corev1.ServiceAccount{
    ObjectMeta: metav1.ObjectMeta{
        Name:      "error-backend-sa",
        Namespace: req.Namespace,
    },
}
if err := r.applyServiceAccount(ctx, serviceAccount); err != nil {
    log.Error(err, "Failed to create/update ServiceAccount")
    return ctrl.Result{}, err
}

// Define the Cluster role resource
clusterRole := &rbacv1.ClusterRole{
    ObjectMeta: metav1.ObjectMeta{
        Name: "error-backend-role",
    },
    Rules: []rbacv1.PolicyRule{
        {
            APIGroups: []string{""},
            Resources: []string{"events", "pods/log"},
            Verbs:     []string{"get", "list", "watch"},
        },
    },
}
if err := r.applyClusterRole(ctx, clusterRole); err != nil {
    log.Error(err, "Failed to create/update ClusterRole")
    return ctrl.Result{}, err
}

// Define the Cluster Role Binding resource
clusterRoleBinding := &rbacv1.ClusterRoleBinding{
    ObjectMeta: metav1.ObjectMeta{
        Name: "error-backend-binding",
    },
    Subjects: []rbacv1.Subject{
        {
            Kind:      "ServiceAccount",
            Name:      "error-backend-sa",
            Namespace: req.Namespace,
        },
    },
    RoleRef: rbacv1.RoleRef{
        Kind:     "ClusterRole",
        Name:     "error-backend-role",
        APIGroup: "rbac.authorization.k8s.io",
    },
}
if err := r.applyClusterRoleBinding(ctx, clusterRoleBinding); err != nil {
    log.Error(err, "Failed to create/update ClusterRoleBinding")
    return ctrl.Result{}, err
}

// Define the MongoDB Deployment
mongoDeployment := &appsv1.Deployment{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "mongodb",
		Namespace: req.Namespace,
	},
	Spec: appsv1.DeploymentSpec{
		Replicas: pointerToInt32(1),
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{"app": "mongodb"},
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{"app": "mongodb"},
			},
			Spec: corev1.PodSpec{
				Containers: []corev1.Container{
					{
						Name:  "mongodb",
						Image: "mongo:latest",
						Ports: []corev1.ContainerPort{{ContainerPort: 27017}},
						VolumeMounts: []corev1.VolumeMount{
							{Name: "mongo-data", MountPath: "/data/db"},
						},
					},
				},
				Volumes: []corev1.Volume{
					{
						Name: "mongo-data",
						VolumeSource: corev1.VolumeSource{
							HostPath: &corev1.HostPathVolumeSource{
								Path: errorkube.Spec.MongoDBHostPath,
								Type: pointerToHostPathType(corev1.HostPathDirectoryOrCreate),
							},
						},
					},
				},
			},
		},
	},
}
if err := controllerutil.SetControllerReference(errorkube, mongoDeployment, r.Scheme); err != nil {
	return ctrl.Result{}, err
}
if err := r.applyDeployment(ctx, mongoDeployment); err != nil {
	log.Error(err, "Failed to create/update MongoDB Deployment")
	return ctrl.Result{}, err
}

// Define the MongoDB Service
mongoService := &corev1.Service{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "mongo-service",
		Namespace: req.Namespace,
	},
	Spec: corev1.ServiceSpec{
		Selector: map[string]string{"app": "mongodb"},
		Ports: []corev1.ServicePort{
			{Protocol: corev1.ProtocolTCP, Port: 27017, TargetPort: intstr.FromInt(27017)},
		},
		Type: corev1.ServiceTypeClusterIP,
	},
}
if err := controllerutil.SetControllerReference(errorkube, mongoService, r.Scheme); err != nil {
	return ctrl.Result{}, err
}
if err := r.applyService(ctx, mongoService); err != nil {
	log.Error(err, "Failed to create/update MongoDB Service")
	return ctrl.Result{}, err
}

// Check MongoDB readiness before creating Backend Deployment
mongoURI := "mongodb://mongo-service.default.svc.cluster.local:27017"
if !r.isMongoDBReady(ctx ,mongoURI) {
	log.Info("Waiting for MongoDB to be ready for communication")
	return ctrl.Result{RequeueAfter: 10 * time.Second}, nil
}


// Define the Backend Deployment
backendDeployment := &appsv1.Deployment{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "error-backend",
		Namespace: req.Namespace,
	},
	Spec: appsv1.DeploymentSpec{
		Replicas: pointerToInt32(errorkube.Spec.Replicas),
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{"app": "error-backend"},
		},
		Template: corev1.PodTemplateSpec{
			ObjectMeta: metav1.ObjectMeta{
				Labels: map[string]string{"app": "error-backend"},
			},
			Spec: corev1.PodSpec{
				ServiceAccountName: "error-backend-sa",
				Containers: []corev1.Container{
					{
						Name:  "error-backend",
						Image: "santhosh9515/errorkube:latest",
						Ports: []corev1.ContainerPort{{ContainerPort: 8080}},
						Env: []corev1.EnvVar{
							{Name: "MONGODB_URI", Value: "mongodb://mongo-service.default.svc.cluster.local:27017"},
						},
					},
				},
			},
		},
	},
}
if err := controllerutil.SetControllerReference(errorkube, backendDeployment, r.Scheme); err != nil {
	return ctrl.Result{}, err
}
if err := r.applyDeployment(ctx, backendDeployment); err != nil {
	log.Error(err, "Failed to create/update Backend Deployment")
	return ctrl.Result{}, err
}

// Define the Backend Service
backendService := &corev1.Service{
	ObjectMeta: metav1.ObjectMeta{
		Name:      "backend-service",
		Namespace: req.Namespace,
	},
	Spec: corev1.ServiceSpec{
		Selector: map[string]string{"app": "error-backend"},
		Ports: []corev1.ServicePort{
			{
				Protocol:   corev1.ProtocolTCP,
				Port:       errorkube.Spec.BackendPort,
				TargetPort: intstr.FromInt(8080),
			},
		},
		Type: corev1.ServiceType(errorkube.Spec.BackendServiceType),
	},
}
if err := controllerutil.SetControllerReference(errorkube, backendService, r.Scheme); err != nil {
	return ctrl.Result{}, err
}
if err := r.applyService(ctx, backendService); err != nil {
	log.Error(err, "Failed to create/update Backend Service")
	return ctrl.Result{}, err
}

return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ErrorKubeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&errorkubev1.ErrorKube{}).
		Complete(r)
}

// Utility functions
func (r *ErrorKubeReconciler) applyServiceAccount(ctx context.Context, sa *corev1.ServiceAccount) error {
    existing := &corev1.ServiceAccount{}
    err := r.Get(ctx, client.ObjectKeyFromObject(sa), existing)
    if err != nil {
        if client.IgnoreNotFound(err) != nil {
            return err
        }
        return r.Create(ctx, sa)
    }
    return nil // ServiceAccount updates are rare; skip update logic
}

func (r *ErrorKubeReconciler) applyClusterRole(ctx context.Context, cr *rbacv1.ClusterRole) error {
    existing := &rbacv1.ClusterRole{}
    err := r.Get(ctx, client.ObjectKeyFromObject(cr), existing)
    if err != nil {
        if client.IgnoreNotFound(err) != nil {
            return err
        }
        return r.Create(ctx, cr)
    }
    return nil // ClusterRole updates are rare; skip update logic
}

func (r *ErrorKubeReconciler) applyClusterRoleBinding(ctx context.Context, crb *rbacv1.ClusterRoleBinding) error {
    existing := &rbacv1.ClusterRoleBinding{}
    err := r.Get(ctx, client.ObjectKeyFromObject(crb), existing)
    if err != nil {
        if client.IgnoreNotFound(err) != nil {
            return err
        }
        return r.Create(ctx, crb)
    }
    return nil // ClusterRoleBinding updates are rare; skip update logic
}


func (r *ErrorKubeReconciler) applyDeployment(ctx context.Context, deployment *appsv1.Deployment) error {
	existing := &appsv1.Deployment{}
	err := r.Get(ctx, client.ObjectKeyFromObject(deployment), existing)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			return err
		}
		return r.Create(ctx, deployment)
	}
	deployment.ResourceVersion = existing.ResourceVersion
	return r.Update(ctx, deployment)
}

func (r *ErrorKubeReconciler) applyService(ctx context.Context, service *corev1.Service) error {
	existing := &corev1.Service{}
	err := r.Get(ctx, client.ObjectKeyFromObject(service), existing)
	if err != nil {
		if client.IgnoreNotFound(err) != nil {
			return err
		}
		return r.Create(ctx, service)
	}
	service.Spec.ClusterIP = existing.Spec.ClusterIP // Preserve ClusterIP during updates
	return r.Update(ctx, service)
}

// Checks if the MongoDB database is ready by connecting to its endpoint.
func (r *ErrorKubeReconciler) isMongoDBReady(ctx context.Context ,mongoURI string) bool {
	log := log.FromContext(ctx)
	clientOptions := options.Client().ApplyURI(mongoURI)
	client, err := mongo.NewClient(clientOptions)
	if err != nil {
		log.Error(err, "Failed to create MongoDB client")
		return false
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		log.Error(err, "Failed to connect to MongoDB")
		return false
	}

	// Check connection status
	if err := client.Ping(ctx, nil); err != nil {
		log.Error(err, "MongoDB is not responding to ping")
		return false
	}

	log.Info("MongoDB is ready")
	return true
}

func pointerToInt32(value int32) *int32 {
	return &value
}

func pointerToHostPathType(value corev1.HostPathType) *corev1.HostPathType {
	return &value
}