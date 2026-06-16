# Architecture

## Overview

k8s-operator-go follows the standard controller-runtime pattern: a single binary runs as a Deployment inside the cluster, watching AppDeployment resources via informers and reconciling desired state.

## Components

### Manager

The Manager (cmd/manager/main.go) bootstraps:
- Shared informers and client cache
- Controller registration
- Webhook server
- Health probes (healthz, readyz)
- Prometheus metrics endpoint

### Controller

The AppDeploymentController watches:
- **Primary**: AppDeployment (the custom resource)
- **Owned**: Deployment, Service (created by the controller)

Reconcile loop:
1. Fetch AppDeployment
2. Handle deletion (finalizer → cleanup)
3. Reconcile Deployment (create/update)
4. Reconcile Service (create/update)
5. Update status with phase and conditions

### Webhooks

- **Mutating**: Sets defaults (replicas=1, strategy=Rolling, port=8080)
- **Validating**: Rejects invalid specs (missing image, bad port, etc.)

### Metrics

All metrics are served on `:8080/metrics` via controller-runtime's built-in HTTP server.

## CRD

The AppDeployment CRD is registered under `api.asier.dev/v1alpha1` with:
- Status subresource for phase tracking
- Printer columns for kubectl UX
- OpenAPI v3 schema validation
