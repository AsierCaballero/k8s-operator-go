package controllers

import (
	"testing"
	"time"

	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	appv1alpha1 "github.com/AsierCaballero/k8s-operator-go/api/v1alpha1"
)

func TestAppDeploymentCreation(t *testing.T) {
	g := NewWithT(t)
	app := testAppDeployment("test-creation", "default", "nginx:1.25")
	mustCreateAppDeployment(t, app)

	deploy := &appsv1.Deployment{}
	g.Eventually(func() error {
		return k8sClient.Get(ctx, types.NamespacedName{Name: app.Name, Namespace: app.Namespace}, deploy)
	}, 5*time.Second, 100*time.Millisecond).Should(Succeed())

	g.Expect(deploy.Spec.Replicas).ToNot(BeNil())
	g.Expect(*deploy.Spec.Replicas).To(Equal(int32(3)))
	g.Expect(deploy.Spec.Template.Spec.Containers).To(HaveLen(1))
	g.Expect(deploy.Spec.Template.Spec.Containers[0].Image).To(Equal("nginx:1.25"))
}

func TestAppDeploymentServiceCreated(t *testing.T) {
	g := NewWithT(t)
	app := testAppDeployment("test-service", "default", "redis:7-alpine")
	mustCreateAppDeployment(t, app)

	svc := &corev1.Service{}
	g.Eventually(func() error {
		return k8sClient.Get(ctx, types.NamespacedName{Name: app.Name, Namespace: app.Namespace}, svc)
	}, 5*time.Second, 100*time.Millisecond).Should(Succeed())

	g.Expect(svc.Spec.Ports).To(HaveLen(1))
	g.Expect(svc.Spec.Ports[0].Port).To(Equal(int32(8080)))
	g.Expect(svc.Spec.Selector).To(HaveKeyWithValue("app.kubernetes.io/instance", app.Name))
}

func TestAppDeploymentReplicasUpdate(t *testing.T) {
	g := NewWithT(t)
	app := testAppDeployment("test-scale", "default", "nginx:1.25")
	mustCreateAppDeployment(t, app)

	deploy := &appsv1.Deployment{}
	g.Eventually(func() error {
		return k8sClient.Get(ctx, types.NamespacedName{Name: app.Name, Namespace: app.Namespace}, deploy)
	}, 5*time.Second, 100*time.Millisecond).Should(Succeed())
	g.Expect(*deploy.Spec.Replicas).To(Equal(int32(3)))

	updated := app.DeepCopy()
	newReplicas := int32(5)
	updated.Spec.Replicas = &newReplicas
	g.Expect(k8sClient.Update(ctx, updated)).To(Succeed())

	g.Eventually(func() int32 {
		if err := k8sClient.Get(ctx, types.NamespacedName{Name: app.Name, Namespace: app.Namespace}, deploy); err != nil {
			return 0
		}
		if deploy.Spec.Replicas == nil {
			return 0
		}
		return *deploy.Spec.Replicas
	}, 5*time.Second, 100*time.Millisecond).Should(Equal(int32(5)))
}

func TestAppDeploymentImageUpdate(t *testing.T) {
	g := NewWithT(t)
	app := testAppDeployment("test-image-update", "default", "nginx:1.25")
	mustCreateAppDeployment(t, app)

	deploy := &appsv1.Deployment{}
	g.Eventually(func() error {
		return k8sClient.Get(ctx, types.NamespacedName{Name: app.Name, Namespace: app.Namespace}, deploy)
	}, 5*time.Second, 100*time.Millisecond).Should(Succeed())

	updated := app.DeepCopy()
	updated.Spec.Image = "nginx:1.27"
	g.Expect(k8sClient.Update(ctx, updated)).To(Succeed())

	g.Eventually(func() string {
		if err := k8sClient.Get(ctx, types.NamespacedName{Name: app.Name, Namespace: app.Namespace}, deploy); err != nil {
			return ""
		}
		return deploy.Spec.Template.Spec.Containers[0].Image
	}, 5*time.Second, 100*time.Millisecond).Should(Equal("nginx:1.27"))
}

func TestAppDeploymentEnvVars(t *testing.T) {
	g := NewWithT(t)
	app := testAppDeployment("test-env", "default", "app:latest")
	mustCreateAppDeployment(t, app)

	deploy := &appsv1.Deployment{}
	g.Eventually(func() error {
		return k8sClient.Get(ctx, types.NamespacedName{Name: app.Name, Namespace: app.Namespace}, deploy)
	}, 5*time.Second, 100*time.Millisecond).Should(Succeed())

	updated := app.DeepCopy()
	updated.Spec.Env = []corev1.EnvVar{
		{Name: "FOO", Value: "bar"},
		{Name: "BAZ", Value: "qux"},
	}
	g.Expect(k8sClient.Update(ctx, updated)).To(Succeed())

	g.Eventually(func() int {
		if err := k8sClient.Get(ctx, types.NamespacedName{Name: app.Name, Namespace: app.Namespace}, deploy); err != nil {
			return -1
		}
		return len(deploy.Spec.Template.Spec.Containers[0].Env)
	}, 5*time.Second, 100*time.Millisecond).Should(Equal(2))
}

func TestAppDeploymentDeletion(t *testing.T) {
	g := NewWithT(t)
	app := testAppDeployment("test-deletion", "default", "nginx:1.25")
	mustCreateAppDeployment(t, app)

	deploy := &appsv1.Deployment{}
	svc := &corev1.Service{}
	g.Eventually(func() error {
		return k8sClient.Get(ctx, types.NamespacedName{Name: app.Name, Namespace: app.Namespace}, deploy)
	}, 5*time.Second, 100*time.Millisecond).Should(Succeed())
	g.Eventually(func() error {
		return k8sClient.Get(ctx, types.NamespacedName{Name: app.Name, Namespace: app.Namespace}, svc)
	}, 5*time.Second, 100*time.Millisecond).Should(Succeed())

	g.Expect(k8sClient.Delete(ctx, app)).To(Succeed())

	g.Eventually(func() error {
		return k8sClient.Get(ctx, types.NamespacedName{Name: app.Name, Namespace: app.Namespace}, deploy)
	}, 5*time.Second, 100*time.Millisecond).ShouldNot(Succeed())
	g.Eventually(func() error {
		return k8sClient.Get(ctx, types.NamespacedName{Name: app.Name, Namespace: app.Namespace}, svc)
	}, 5*time.Second, 100*time.Millisecond).ShouldNot(Succeed())
}

func TestAppDeploymentStatus(t *testing.T) {
	g := NewWithT(t)
	app := testAppDeployment("test-status", "default", "nginx:1.25")
	mustCreateAppDeployment(t, app)

	g.Eventually(func() int64 {
		latest := &appv1alpha1.AppDeployment{}
		if err := k8sClient.Get(ctx, types.NamespacedName{Name: app.Name, Namespace: app.Namespace}, latest); err != nil {
			return 0
		}
		return latest.Status.ObservedGeneration
	}, 5*time.Second, 100*time.Millisecond).Should(BeNumerically(">=", 1))
}

func TestAppDeploymentWithCustomStrategy(t *testing.T) {
	g := NewWithT(t)
	app := testAppDeployment("test-strategy", "default", "nginx:1.25")
	app.Spec.Strategy = appv1alpha1.StrategyRecreate
	mustCreateAppDeployment(t, app)

	deploy := &appsv1.Deployment{}
	g.Eventually(func() error {
		return k8sClient.Get(ctx, types.NamespacedName{Name: app.Name, Namespace: app.Namespace}, deploy)
	}, 5*time.Second, 100*time.Millisecond).Should(Succeed())
	g.Expect(string(deploy.Spec.Strategy.Type)).To(Equal("Recreate"))
}

func TestAppDeploymentDeploymentLabels(t *testing.T) {
	g := NewWithT(t)
	app := testAppDeployment("test-labels", "default", "nginx:1.25")
	mustCreateAppDeployment(t, app)

	deploy := &appsv1.Deployment{}
	g.Eventually(func() error {
		return k8sClient.Get(ctx, types.NamespacedName{Name: app.Name, Namespace: app.Namespace}, deploy)
	}, 5*time.Second, 100*time.Millisecond).Should(Succeed())

	g.Expect(deploy.Labels).To(HaveKeyWithValue("app.kubernetes.io/managed-by", "k8s-operator-go"))
	g.Expect(deploy.Labels).To(HaveKeyWithValue("app.kubernetes.io/instance", app.Name))
}

func TestAppDeploymentServicePorts(t *testing.T) {
	g := NewWithT(t)
	app := testAppDeployment("test-svc-ports", "default", "nginx:1.25")
	app.Spec.Port = 3000
	mustCreateAppDeployment(t, app)

	svc := &corev1.Service{}
	g.Eventually(func() error {
		return k8sClient.Get(ctx, types.NamespacedName{Name: app.Name, Namespace: app.Namespace}, svc)
	}, 5*time.Second, 100*time.Millisecond).Should(Succeed())

	g.Expect(svc.Spec.Ports[0].Port).To(Equal(int32(3000)))
	g.Expect(svc.Spec.Ports[0].TargetPort.IntVal).To(Equal(int32(3000)))
}

func TestAppDeploymentDefaultReplicas(t *testing.T) {
	g := NewWithT(t)
	app := &appv1alpha1.AppDeployment{
		ObjectMeta: metav1.ObjectMeta{Name: "test-default-replicas", Namespace: "default"},
		Spec:       appv1alpha1.AppDeploymentSpec{Image: "nginx:1.25", Port: 80},
	}
	mustCreateAppDeployment(t, app)

	deploy := &appsv1.Deployment{}
	g.Eventually(func() error {
		return k8sClient.Get(ctx, types.NamespacedName{Name: app.Name, Namespace: app.Namespace}, deploy)
	}, 5*time.Second, 100*time.Millisecond).Should(Succeed())

	g.Expect(deploy.Spec.Replicas).ToNot(BeNil())
	g.Expect(*deploy.Spec.Replicas).To(Equal(int32(1)))
}

func TestAppDeploymentResourceLimits(t *testing.T) {
	g := NewWithT(t)
	app := testAppDeployment("test-resources", "default", "nginx:1.25")
	app.Spec.Resources = corev1.ResourceRequirements{
		Requests: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("100m"),
			corev1.ResourceMemory: resource.MustParse("128Mi"),
		},
		Limits: corev1.ResourceList{
			corev1.ResourceCPU:    resource.MustParse("500m"),
			corev1.ResourceMemory: resource.MustParse("256Mi"),
		},
	}
	mustCreateAppDeployment(t, app)

	deploy := &appsv1.Deployment{}
	g.Eventually(func() error {
		return k8sClient.Get(ctx, types.NamespacedName{Name: app.Name, Namespace: app.Namespace}, deploy)
	}, 5*time.Second, 100*time.Millisecond).Should(Succeed())

	req := deploy.Spec.Template.Spec.Containers[0].Resources.Requests
	lim := deploy.Spec.Template.Spec.Containers[0].Resources.Limits
	g.Expect(req.Cpu().Equal(resource.MustParse("100m"))).To(BeTrue())
	g.Expect(req.Memory().Equal(resource.MustParse("128Mi"))).To(BeTrue())
	g.Expect(lim.Cpu().Equal(resource.MustParse("500m"))).To(BeTrue())
	g.Expect(lim.Memory().Equal(resource.MustParse("256Mi"))).To(BeTrue())
}
