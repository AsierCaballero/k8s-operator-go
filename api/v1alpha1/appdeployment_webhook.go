package v1alpha1

import (
	"fmt"
	"regexp"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

var appdeploymentLog = logf.Log.WithName("appdeployment-webhook")

var validImagePattern = regexp.MustCompile(`^[a-zA-Z0-9./_-]+(:[a-zA-Z0-9._-]+)?$`)

func (a *AppDeployment) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(a).
		Complete()
}

//+kubebuilder:webhook:path=/mutate-api-asier-dev-v1alpha1-appdeployment,mutating=true,failurePolicy=fail,sideEffects=None,groups=api.asier.dev,resources=appdeployments,verbs=create;update,versions=v1alpha1,name=mappdeployment.api.asier.dev,admissionReviewVersions=v1

var _ webhook.Defaulter = &AppDeployment{}

func (a *AppDeployment) Default() {
	appdeploymentLog.Info("defaulting", "name", a.Name)

	if a.Spec.Replicas == nil || *a.Spec.Replicas < 1 {
		defaultReplicas := int32(1)
		a.Spec.Replicas = &defaultReplicas
	}

	if a.Spec.Strategy == "" {
		a.Spec.Strategy = StrategyRolling
	}

	if a.Spec.Port == 0 {
		a.Spec.Port = 8080
	}

	if a.Spec.Ingress != nil && a.Spec.Ingress.Path == "" {
		a.Spec.Ingress.Path = "/"
	}

	if a.Annotations == nil {
		a.Annotations = make(map[string]string)
	}
	if _, ok := a.Annotations["api.asier.dev/last-applied"]; !ok {
		a.Annotations["api.asier.dev/last-applied"] = fmt.Sprintf("%s:%s", a.Spec.Image, a.Spec.Strategy)
	}
}

//+kubebuilder:webhook:path=/validate-api-asier-dev-v1alpha1-appdeployment,mutating=false,failurePolicy=fail,sideEffects=None,groups=api.asier.dev,resources=appdeployments,verbs=create;update,versions=v1alpha1,name=vappdeployment.api.asier.dev,admissionReviewVersions=v1

var _ webhook.Validator = &AppDeployment{}

func (a *AppDeployment) ValidateCreate() (admission.Warnings, error) {
	appdeploymentLog.Info("validate create", "name", a.Name)
	return a.validateAppDeployment()
}

func (a *AppDeployment) ValidateUpdate(old runtime.Object) (admission.Warnings, error) {
	appdeploymentLog.Info("validate update", "name", a.Name)
	return a.validateAppDeployment()
}

func (a *AppDeployment) ValidateDelete() (admission.Warnings, error) {
	return nil, nil
}

func (a *AppDeployment) validateAppDeployment() (admission.Warnings, error) {
	var warnings admission.Warnings

	if a.Spec.Image == "" {
		return warnings, fmt.Errorf("spec.image is required")
	}

	if !validImagePattern.MatchString(a.Spec.Image) {
		warnings = append(warnings, fmt.Sprintf("image %q does not match expected pattern", a.Spec.Image))
	}

	if a.Spec.Port < 1 || a.Spec.Port > 65535 {
		return warnings, fmt.Errorf("spec.port must be between 1 and 65535, got %d", a.Spec.Port)
	}

	if a.Spec.Replicas != nil && *a.Spec.Replicas < 1 {
		return warnings, fmt.Errorf("spec.replicas must be >= 1, got %d", *a.Spec.Replicas)
	}

	switch a.Spec.Strategy {
	case StrategyRolling, StrategyRecreate, StrategyBlueGreen:
	case "":
	default:
		return warnings, fmt.Errorf("unsupported strategy %q, must be one of: Rolling, Recreate, BlueGreen", a.Spec.Strategy)
	}

	if a.Spec.Canary != nil {
		totalWeight := int32(0)
		for i, step := range a.Spec.Canary.Steps {
			if step.Weight < 0 || step.Weight > 100 {
				return warnings, fmt.Errorf("canary step %d: weight must be between 0 and 100", i)
			}
			totalWeight += step.Weight
		}
		if totalWeight > 100 {
			warnings = append(warnings, fmt.Sprintf("canary steps total weight %d%% exceeds 100%%", totalWeight))
		}
	}

	if a.Spec.Ingress != nil && a.Spec.Ingress.Enabled {
		if a.Spec.Ingress.Host == "" {
			warnings = append(warnings, "ingress is enabled but host is empty")
		}
	}

	if a.Spec.Resources.Requests != nil || a.Spec.Resources.Limits != nil {
		warnings = append(warnings, "resource constraints set — ensure cluster has sufficient capacity")
	}

	return warnings, nil
}
