package sosreport

import (
	"context"
	"fmt"
	sosv1alpha1 "github.com/bostrt/sosreport-operator/pkg/apis/sos/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/uuid"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var (
	log = logf.Log.WithName("controller_sosreport")
	sosreportDestDir = "/host/var/tmp/sosreports"
	sosreportDestAnnotation = "destination-dir"
)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Sosreport Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileSosreport{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("sosreport-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Sosreport
	err = c.Watch(&source.Kind{Type: &sosv1alpha1.Sosreport{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Sosreport
	err = c.Watch(&source.Kind{Type: &batchv1.Job{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &sosv1alpha1.Sosreport{},
	})

	// TODO: Watch for Pods to update SosreportStatus? Can we just watch Job?
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileSosreport implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileSosreport{}

// ReconcileSosreport reconciles a Sosreport object
type ReconcileSosreport struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Sosreport object and makes changes based on the state read
// and what is in the Sosreport.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileSosreport) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Sosreport")

	// Fetch the Sosreport instance
	instance := &sosv1alpha1.Sosreport{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Define a new Job object
	job := newJobForCR(instance, "ip-10-0-129-182")

	// Set Sosreport instance as the owner and controller
	if err := controllerutil.SetControllerReference(instance, job, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check if this Job already exists
	found := &batchv1.Job{}

	err = r.client.Get(context.TODO(), types.NamespacedName{Name: job.Name, Namespace: job.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Creating a new Job", "Job.Namespace", job.Namespace, "Job.Name", job.Name)
		err = r.client.Create(context.TODO(), job)
		if err != nil {
			return reconcile.Result{}, err
		}

		// Job created successfully - don't requeue
		return reconcile.Result{}, nil
	} else if err != nil {
		return reconcile.Result{}, err
	}

	// Update SosreportStatus if job has completed.
	if found.Status.Succeeded > 0 {
		// Extract controller-uid from Job and match against Pods
		jobControllerUid := found.Spec.Selector.MatchLabels["controller-uid"]
		pods := &corev1.PodList{}
		opts := []client.ListOption{
			client.InNamespace(request.NamespacedName.Namespace),
			client.MatchingLabels{"controller-uid": jobControllerUid},
		}
		err = r.client.List(context.TODO(), pods, opts...)
		if err != nil {
			return reconcile.Result{}, err
		}

		//  TODO: Do something useful here...
		for _, p := range pods.Items {
			reqLogger.Info(p.Name)

		}
	}
	// Job already exists - don't requeue
	reqLogger.Info("Skip reconcile: Job already exists", "Job.Namespace", found.Namespace, "Job.Name", found.Name)
	return reconcile.Result{}, nil
}

func newJobsForCR(cr *sosv1alpha1.Sosreport) []*batchv1.Job {
	return nil
}

func newJobForCR(cr *sosv1alpha1.Sosreport, nodeName string) *batchv1.Job {
	jobUuid := string(uuid.NewUUID())
	jobDestinationDirectory := fmt.Sprintf("%s/%s", sosreportDestDir, jobUuid)
	jobName := fmt.Sprintf("%s-job", cr.Name)
	priv := true

	labels := map[string]string{
		"app": cr.Name, // TODO: Fix name
	}
	annotations := map[string]string{
		sosreportDestAnnotation: jobDestinationDirectory,
	}
	nodeSelector := map[string]string{
		"kubernetes.io/hostname": nodeName,
	}
	return &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name: jobName,
			Namespace: cr.Namespace,
			Labels: labels,
			Annotations: annotations,
		},
		Spec: batchv1.JobSpec{
				Template: corev1.PodTemplateSpec{
					ObjectMeta: metav1.ObjectMeta{},
					Spec:       corev1.PodSpec{
						RestartPolicy: "Never",
						HostNetwork: true,
						HostPID: true,
						ServiceAccountName: "sosreport-worker",
						Containers: []corev1.Container{
							{
								Name: "sosreport-worker", // TODO: Fix name
								Image: "quay.io/bostrt/sosreport-worker",
								Args: []string{jobDestinationDirectory, "-o", "lvm2"},
								SecurityContext: &corev1.SecurityContext{
									Privileged: &priv,
								},
								VolumeMounts: []corev1.VolumeMount{
									{
										Name: "host",
										MountPath: "/host",
									},
								},
							},
						},
						NodeSelector: nodeSelector,
						Tolerations: []corev1.Toleration{
							{
								Operator: "Exists",
							},
						},
						Volumes: []corev1.Volume{
							{
								Name: "host",
								VolumeSource: corev1.VolumeSource{
									HostPath: &corev1.HostPathVolumeSource{
										Path: "/",
										Type: nil,
									},
								},
							},
						},
					},
			},
		},
	}
}