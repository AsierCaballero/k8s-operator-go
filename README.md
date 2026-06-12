# k8s-operator-go

[![CI](https://img.shields.io/github/actions/workflow/status/AsierCaballero/k8s-operator-go/ci.yml?label=CI&logo=github)](https://github.com/AsierCaballero/k8s-operator-go/actions)
[![Go](https://img.shields.io/badge/go-1.22+-00ADD8?logo=go)](https://go.dev)
[![controller-runtime](https://img.shields.io/badge/controller--runtime-v0.18-blue)](https://github.com/kubernetes-sigs/controller-runtime)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen)](CONTRIBUTING.md)

A Kubernetes operator for declarative application deployments. Manage applications via a custom resource — the operator handles Deployment, Service, canary rollouts, and self-healing.

```bash
kubectl apply -f - <<EOF
apiVersion: api.asier.dev/v1alpha1
kind: AppDeployment
metadata:
  name: myapp
  namespace: default
spec:
  image: nginx:1.25
  replicas: 3
  port: 80
EOF
```

---

## Features

- **Declarative deployments** — define apps via `AppDeployment` CRD
- **Automatic lifecycle** — creates and manages Deployment + Service
- **Update strategies** — Rolling (default), Recreate, BlueGreen
- **Canary support** — weighted traffic splits for progressive delivery
- **Self-healing** — controller converges desired state automatically
- **Status reporting** — phase, conditions, observed generation
- **Admission webhooks** — validation and defaulting on CR create/update
- **Prometheus metrics** — reconcile duration, errors, resource ops, phase tracking
- **Finalizer cleanup** — garbage collection of child resources on deletion

## Quick start

### Prerequisites

- Kubernetes cluster 1.27+
- `kubectl` configured

### Install the operator

```bash
make install
make deploy
```

### Create an application

```bash
kubectl apply -f config/samples/appdeployment_sample.yaml
```

### Verify

```bash
kubectl get appdeployments
kubectl get pods -l app.kubernetes.io/instance=webapp-sample
```

## Configuration

| Spec field | Type | Default | Description |
|---|---|---|---|
| `image` | string | — | Container image (required) |
| `replicas` | int32 | 1 | Number of replicas |
| `port` | int32 | 8080 | Container port |
| `strategy` | string | Rolling | Update strategy |
| `env` | EnvVar[] | — | Environment variables |
| `resources` | ResourceRequirements | — | CPU/memory requests and limits |
| `canary` | CanaryConfig | — | Canary rollout configuration |
| `ingress` | IngressConfig | — | Ingress settings |

## Architecture

```
┌──────────────────────────────────────────────┐
│                 Kubernetes                    │
│                                               │
│  ┌──────────────┐   watches   ┌────────────┐ │
│  │  AppDeployment │◄──────────│ Controller  │ │
│  │    (CRD)      │            │ (Reconciler)│ │
│  └──────────────┘            └──────┬───────┘ │
│                                     │         │
│                            ┌────────┴───────┐ │
│                            │  creates/owns   │ │
│                            └──┬──────────┬──┘ │
│                      ┌────────┴──┐ ┌─────┴──┐ │
│                      │ Deployment│ │ Service │ │
│                      │ (Pods)    │ │ (Net)   │ │
│                      └───────────┘ └────────┘ │
└──────────────────────────────────────────────┘
         ┌──────────────┐      ┌──────────────┐
         │  Admission   │      │  Prometheus   │
         │  Webhooks    │      │   Metrics     │
         └──────────────┘      └──────────────┘
```

## Development

### Prerequisites

- Go 1.22+
- kubebuilder or envtest binaries for local testing

### Setup

```bash
git clone https://github.com/AsierCaballero/k8s-operator-go.git
cd k8s-operator-go
make generate       # Generate deepcopy + CRD manifests
make test           # Run tests (requires envtest)
make build          # Compile the manager binary
```

### Running locally

```bash
make run
```

## Project structure

```
├── api/v1alpha1/           # CRD types and webhooks
├── cmd/manager/            # Entry point
├── config/
│   ├── crd/                # CRD manifest
│   ├── manager/            # Operator Deployment manifest
│   ├── rbac/               # RBAC configuration
│   ├── samples/            # Example AppDeployments
│   └── webhook/            # Webhook configuration
├── controllers/            # Reconciler and tests
├── docs/                   # Documentation
└── internal/metrics/       # Prometheus metrics
```

## Author

**Asier Caballero** — Senior DevOps Engineer & Cloud Architect
asier.caballero1@gmail.com · [linkedin.com/in/asier-caballero](https://linkedin.com/in/asier-caballero)
