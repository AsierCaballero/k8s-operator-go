package controllers

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	appv1alpha1 "github.com/AsierCaballero/k8s-operator-go/api/v1alpha1"
	"github.com/AsierCaballero/k8s-operator-go/internal/metrics"
)

//+kubebuilder:rbac:groups=api.asier.dev,resources=appdeployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=api.asier.dev,resources=appdeployments/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=api.asier.dev,resources=appdeployments/finalizers,verbs=update
//+kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get
//+kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=core,resources=events,verbs=create;patch
//+kubebuilder:rbac:groups=coordination.k8s.io,resources=leases,verbs=get;list;watch;create;update;patch;delete

const (
	appDeploymentFinalizer = "api.asier.dev/finalizer"
	deploymentHashLabel    = "api.asier.dev/deployment-hash"
	appNameLabel           = "app.kubernetes.io/name"
	appInstanceLabel       = "app.kubernetes.io/instance"
	appManagedByLabel      = "app.kubernetes.io/managed-by"
)

type AppDeploymentReconciler struct {
	client.Client
	Scheme  *runtime.Scheme
	Metrics *metrics.OperatorMetrics
}

func (r *AppDeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appv1alpha1.AppDeployment{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Complete(r)
}

func (r *AppDeploymentReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	start := time.Now()

	app := &appv1alpha1.AppDeployment{}
	if err := r.Get(ctx, req.NamespacedName, app); err != nil {
		if apierrors.IsNotFound(err) {
			r.Metrics.ResetDeploymentPhases(req.Name, req.Namespace)
			return ctrl.Result{}, nil
		}
		r.Metrics.RecordReconcileError("fetch_error")
		return ctrl.Result{}, err
	}

	r.Metrics.AppDeploymentsTotal.Set(float64(len(app.Spec.Env) + 1))
	r.Metrics.RecordResourceOperation("AppDeployment", "get")

	if app.ObjectMeta.DeletionTimestamp.IsZero() {
		if !controllerutil.ContainsFinalizer(app, appDeploymentFinalizer) {
			controllerutil.AddFinalizer(app, appDeploymentFinalizer)
			if err := r.Update(ctx, app); err != nil {
				r.Metrics.RecordReconcileError("finalizer_add_error")
				return ctrl.Result{}, err
			}
			r.Metrics.RecordResourceOperation("AppDeployment", "update_finalizer")
		}
	} else {
		result, err := r.handleCleanup(ctx, app)
		duration := time.Since(start).Seconds()
		r.Metrics.ObserveReconcile("deleted", duration)
		r.Metrics.ResetDeploymentPhases(app.Name, app.Namespace)
		return result, err
	}

	if err := r.reconcileDeployment(ctx, app); err != nil {
		logger.Error(err, "failed to reconcile Deployment")
		r.setCondition(ctx, app, appv1alpha1.PhaseFailed, "DeploymentReconcileError", err.Error())
		r.Metrics.RecordReconcileError("deployment_error")
		r.Metrics.ObserveReconcile("error", time.Since(start).Seconds())
		return ctrl.Result{}, err
	}

	if err := r.reconcileService(ctx, app); err != nil {
		logger.Error(err, "failed to reconcile Service")
		r.setCondition(ctx, app, appv1alpha1.PhaseFailed, "ServiceReconcileError", err.Error())
		r.Metrics.RecordReconcileError("service_error")
		r.Metrics.ObserveReconcile("error", time.Since(start).Seconds())
		return ctrl.Result{}, err
	}

	if err := r.updateStatus(ctx, app); err != nil {
		r.Metrics.RecordReconcileError("status_update_error")
		r.Metrics.ObserveReconcile("error", time.Since(start).Seconds())
		return ctrl.Result{}, err
	}

	r.Metrics.SetDeploymentPhase(app.Name, app.Namespace, string(app.Status.Phase))
	r.Metrics.ObserveReconcile("success", time.Since(start).Seconds())

	return ctrl.Result{}, nil
}

func (r *AppDeploymentReconciler) handleCleanup(ctx context.Context, app *appv1alpha1.AppDeployment) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("handling AppDeployment deletion", "name", app.CanonicalName())

	if controllerutil.ContainsFinalizer(app, appDeploymentFinalizer) {
		if err := r.cleanupResources(ctx, app); err != nil {
			return ctrl.Result{}, err
		}

		controllerutil.RemoveFinalizer(app, appDeploymentFinalizer)
		if err := r.Update(ctx, app); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

func (r *AppDeploymentReconciler) cleanupResources(ctx context.Context, app *appv1alpha1.AppDeployment) error {
	propagation := metav1.DeletePropagationForeground
	policy := client.DeleteOptions{PropagationPolicy: &propagation}

	deploy := &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{
		Name:      app.Name,
		Namespace: app.Namespace,
	}}
	if err := r.Delete(ctx, deploy, &policy); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete Deployment: %w", err)
	}

	svc := &corev1.Service{ObjectMeta: metav1.ObjectMeta{
		Name:      app.Name,
		Namespace: app.Namespace,
	}}
	if err := r.Delete(ctx, svc, &policy); err != nil && !apierrors.IsNotFound(err) {
		return fmt.Errorf("failed to delete Service: %w", err)
	}

	return nil
}

func (r *AppDeploymentReconciler) reconcileDeployment(ctx context.Context, app *appv1alpha1.AppDeployment) error {
	desired := r.buildDeployment(app)

	existing := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: app.Name, Namespace: app.Namespace}, existing)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
		if err := controllerutil.SetControllerReference(app, desired, r.Scheme); err != nil {
			return err
		}
		return r.Create(ctx, desired)
	}

	if !deploymentSpecsEqual(existing, desired) {
		desired.Spec.DeepCopyInto(&existing.Spec)
		existing.Labels = desired.Labels
		if existing.Annotations == nil {
			existing.Annotations = make(map[string]string)
		}
		for k, v := range desired.Annotations {
			existing.Annotations[k] = v
		}
		return r.Update(ctx, existing)
	}

	return nil
}

func (r *AppDeploymentReconciler) buildDeployment(app *appv1alpha1.AppDeployment) *appsv1.Deployment {
	replicas := int32(1)
	if app.Spec.Replicas != nil && *app.Spec.Replicas > 0 {
		replicas = *app.Spec.Replicas
	}

	labels := map[string]string{
		appNameLabel:      "appdeployment",
		appInstanceLabel:   app.Name,
		appManagedByLabel: "k8s-operator-go",
	}

	strategy := appsv1.DeploymentStrategy{}
	switch app.Spec.Strategy {
	case appv1alpha1.StrategyRecreate:
		strategy.Type = appsv1.RecreateDeploymentStrategyType
	default:
		strategy.Type = appsv1.RollingUpdateDeploymentStrategyType
		strategy.RollingUpdate = &appsv1.RollingUpdateDeployment{
			MaxUnavailable: &intstr.FromInt32(25),
			MaxSurge:       &intstr.FromInt32(25),
		}
	}

	container := corev1.Container{
		Name:      app.Name,
		Image:     app.Spec.Image,
		Ports:     []corev1.ContainerPort{{ContainerPort: app.Spec.Port}},
		Env:       app.Spec.Env,
		Resources: app.Spec.Resources,
		ReadinessProbe: &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				TCPSocket: &corev1.TCPSocketAction{
					Port: intstr.FromInt32(app.Spec.Port),
				},
			},
			InitialDelaySeconds: 5,
			PeriodSeconds:       10,
		},
		LivenessProbe: &corev1.Probe{
			ProbeHandler: corev1.ProbeHandler{
				TCPSocket: &corev1.TCPSocketAction{
					Port: intstr.FromInt32(app.Spec.Port),
				},
			},
			InitialDelaySeconds: 15,
			PeriodSeconds:       20,
		},
	}

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:        app.Name,
			Namespace:   app.Namespace,
			Labels:      labels,
			Annotations: map[string]string{"api.asier.dev/owner": app.Name},
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &replicas,
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{appInstanceLabel: app.Name}},
			Strategy: strategy,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{container},
				},
			},
		},
	}
}

func (r *AppDeploymentReconciler) reconcileService(ctx context.Context, app *appv1alpha1.AppDeployment) error {
	desired := r.buildService(app)

	existing := &corev1.Service{}
	err := r.Get(ctx, types.NamespacedName{Name: app.Name, Namespace: app.Namespace}, existing)
	if err != nil {
		if !apierrors.IsNotFound(err) {
			return err
		}
		if err := controllerutil.SetControllerReference(app, desired, r.Scheme); err != nil {
			return err
		}
		return r.Create(ctx, desired)
	}

	if !serviceSpecsEqual(existing, desired) {
		existing.Spec.Ports = desired.Spec.Ports
		existing.Spec.Selector = desired.Spec.Selector
		if existing.Annotations == nil {
			existing.Annotations = make(map[string]string)
		}
		for k, v := range desired.Annotations {
			existing.Annotations[k] = v
		}
		return r.Update(ctx, existing)
	}

	return nil
}

func (r *AppDeploymentReconciler) buildService(app *appv1alpha1.AppDeployment) *corev1.Service {
	labels := map[string]string{
		appNameLabel:      "appdeployment",
		appInstanceLabel:   app.Name,
		appManagedByLabel: "k8s-operator-go",
	}

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      app.Name,
			Namespace: app.Namespace,
			Labels:    labels,
		},
		Spec: corev1.ServiceSpec{
			Selector: map[string]string{appInstanceLabel: app.Name},
			Ports: []corev1.ServicePort{
				{
					Name:       "http",
					Protocol:   corev1.ProtocolTCP,
					Port:       app.Spec.Port,
					TargetPort: intstr.FromInt32(app.Spec.Port),
				},
			},
		},
	}
}

func (r *AppDeploymentReconciler) updateStatus(ctx context.Context, app *appv1alpha1.AppDeployment) error {
	deploy := &appsv1.Deployment{}
	err := r.Get(ctx, types.NamespacedName{Name: app.Name, Namespace: app.Namespace}, deploy)
	if err != nil {
		return err
	}

	app.Status.AvailableReplicas = deploy.Status.AvailableReplicas
	app.Status.ReadyReplicas = deploy.Status.ReadyReplicas
	app.Status.ObservedGeneration = app.Generation

	switch {
	case deploy.Status.ReadyReplicas >= *deploy.Spec.Replicas:
		app.Status.Phase = appv1alpha1.PhaseReady
		r.setCondition(ctx, app, appv1alpha1.PhaseReady, "DeploymentReady", "All replicas are ready")
	case deploy.Status.ReadyReplicas > 0:
		app.Status.Phase = appv1alpha1.PhaseRolling
		r.setCondition(ctx, app, appv1alpha1.PhaseRolling, "Deploying", "Rolling update in progress")
	default:
		app.Status.Phase = appv1alpha1.PhaseDeploying
		r.setCondition(ctx, app, appv1alpha1.PhaseDeploying, "Deploying", "Waiting for replicas to become ready")
	}

	return r.Status().Update(ctx, app)
}

func (r *AppDeploymentReconciler) setCondition(ctx context.Context, app *appv1alpha1.AppDeployment, phase appv1alpha1.AppDeploymentPhase, reason, message string) {
	condition := metav1.Condition{
		Type:               string(phase),
		Status:             metav1.ConditionTrue,
		Reason:             reason,
		Message:            message,
		ObservedGeneration: app.Generation,
		LastTransitionTime: metav1.Now(),
	}
	meta.SetStatusCondition(&app.Status.Conditions, condition)
}
