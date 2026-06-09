package controllers

import (
	"context"
	"path/filepath"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	appv1alpha1 "github.com/AsierCaballero/k8s-operator-go/api/v1alpha1"
)

var k8sClient ctrl.Client
var testEnv *envtest.Environment
var ctx context.Context
var cancel context.CancelFunc

func TestMain(m *testing.M) {
	logf.SetLogger(zap.New(zap.UseDevMode(true)))

	ctx, cancel = context.WithCancel(context.TODO())
	defer cancel()

	testEnv = &envtest.Environment{
		CRDDirectoryPaths:     []string{filepath.Join("..", "config", "crd")},
		ErrorIfCRDPathMissing: true,
	}

	cfg, err := testEnv.Start()
	if err != nil {
		panic(err)
	}
	defer testEnv.Stop()

	if err := appv1alpha1.AddToScheme(scheme.Scheme); err != nil {
		panic(err)
	}
	if err := appsv1.AddToScheme(scheme.Scheme); err != nil {
		panic(err)
	}
	if err := corev1.AddToScheme(scheme.Scheme); err != nil {
		panic(err)
	}

	k8sClient, err = ctrl.NewClient(cfg, ctrl.Options{Scheme: scheme.Scheme})
	if err != nil {
		panic(err)
	}

	m.Run()
}

func mustCreateAppDeployment(t *testing.T, app *appv1alpha1.AppDeployment) {
	t.Helper()
	if err := k8sClient.Create(ctx, app); err != nil {
		t.Fatalf("failed to create AppDeployment: %v", err)
	}
}

func mustGetDeployment(t *testing.T, name, namespace string) *appsv1.Deployment {
	t.Helper()
	deploy := &appsv1.Deployment{}
	if err := k8sClient.Get(ctx, types.NamespacedName{Name: name, Namespace: namespace}, deploy); err != nil {
		t.Fatalf("failed to get Deployment: %v", err)
	}
	return deploy
}

func testAppDeployment(name, namespace, image string) *appv1alpha1.AppDeployment {
	replicas := int32(3)
	return &appv1alpha1.AppDeployment{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: namespace},
		Spec: appv1alpha1.AppDeploymentSpec{
			Image:    image,
			Replicas: &replicas,
			Port:     8080,
			Strategy: appv1alpha1.StrategyRolling,
		},
	}
}
