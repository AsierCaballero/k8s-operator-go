package controllers

import (
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
)

func deploymentSpecsEqual(a, b *appsv1.Deployment) bool {
	if a == nil || b == nil {
		return a == b
	}
	if a.Spec.Replicas != nil && b.Spec.Replicas != nil {
		if *a.Spec.Replicas != *b.Spec.Replicas {
			return false
		}
	} else if a.Spec.Replicas != b.Spec.Replicas {
		return false
	}
	if a.Spec.Template.Spec.Containers[0].Image != b.Spec.Template.Spec.Containers[0].Image {
		return false
	}
	if !reflect.DeepEqual(a.Spec.Template.Spec.Containers[0].Env, b.Spec.Template.Spec.Containers[0].Env) {
		return false
	}
	if !reflect.DeepEqual(a.Spec.Template.Spec.Containers[0].Resources, b.Spec.Template.Spec.Containers[0].Resources) {
		return false
	}
	if a.Spec.Strategy.Type != b.Spec.Strategy.Type {
		return false
	}
	return true
}

func serviceSpecsEqual(a, b *corev1.Service) bool {
	if a == nil || b == nil {
		return a == b
	}
	if !reflect.DeepEqual(a.Spec.Ports, b.Spec.Ports) {
		return false
	}
	if !reflect.DeepEqual(a.Spec.Selector, b.Spec.Selector) {
		return false
	}
	return true
}
