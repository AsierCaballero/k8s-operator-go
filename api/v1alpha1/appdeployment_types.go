package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type AppDeploymentPhase string

const (
	PhasePending   AppDeploymentPhase = "Pending"
	PhaseDeploying AppDeploymentPhase = "Deploying"
	PhaseReady     AppDeploymentPhase = "Ready"
	PhaseRolling   AppDeploymentPhase = "Rolling"
	PhaseFailed    AppDeploymentPhase = "Failed"
)

type UpdateStrategy string

const (
	StrategyRolling   UpdateStrategy = "Rolling"
	StrategyRecreate  UpdateStrategy = "Recreate"
	StrategyBlueGreen UpdateStrategy = "BlueGreen"
)

type CanaryConfig struct {
	Steps          []CanaryStep `json:"steps,omitempty"`
	TrafficIngress bool         `json:"trafficIngress,omitempty"`
}

type CanaryStep struct {
	Weight   int32 `json:"weight"`
	Replicas int32 `json:"replicas,omitempty"`
}

type IngressConfig struct {
	Enabled     bool              `json:"enabled"`
	Host        string            `json:"host,omitempty"`
	Path        string            `json:"path,omitempty"`
	Annotations map[string]string `json:"annotations,omitempty"`
}

type AppDeploymentSpec struct {
	Image     string                          `json:"image"`
	Replicas  *int32                          `json:"replicas,omitempty"`
	Port      int32                           `json:"port"`
	Env       []corev1.EnvVar                 `json:"env,omitempty"`
	Resources corev1.ResourceRequirements     `json:"resources,omitempty"`
	Strategy  UpdateStrategy                  `json:"strategy,omitempty"`
	Canary    *CanaryConfig                   `json:"canary,omitempty"`
	Ingress   *IngressConfig                  `json:"ingress,omitempty"`
}

type AppDeploymentStatus struct {
	AvailableReplicas  int32              `json:"availableReplicas,omitempty"`
	ReadyReplicas      int32              `json:"readyReplicas,omitempty"`
	Phase              AppDeploymentPhase `json:"phase,omitempty"`
	Message            string             `json:"message,omitempty"`
	ObservedGeneration int64              `json:"observedGeneration,omitempty"`
	Conditions         []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:printcolumn:name="Image",type=string,JSONPath=`.spec.image`
// +kubebuilder:printcolumn:name="Replicas",type=integer,JSONPath=`.spec.replicas`
// +kubebuilder:printcolumn:name="Ready",type=integer,JSONPath=`.status.readyReplicas`
// +kubebuilder:printcolumn:name="Phase",type=string,JSONPath=`.status.phase`
// +kubebuilder:printcolumn:name="Age",type=date,JSONPath=`.metadata.creationTimestamp`

type AppDeployment struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              AppDeploymentSpec   `json:"spec,omitempty"`
	Status            AppDeploymentStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

type AppDeploymentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []AppDeployment `json:"items"`
}

func (a *AppDeployment) CanonicalName() string {
	return a.Namespace + "/" + a.Name
}
