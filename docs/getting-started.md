# Getting started

## Install

### 1. Install CRDs

```bash
make install
```

### 2. Deploy the operator

```bash
make deploy
```

### 3. Verify

```bash
kubectl get pods -n k8s-operator-system
kubectl get crd appdeployments.api.asier.dev
```

## Create your first app

```bash
kubectl apply -f config/samples/appdeployment_sample.yaml
```

Check the status:

```bash
kubectl get appdeployments
kubectl describe appdeployment webapp-sample
```

## Update an app

```bash
kubectl patch appdeployment webapp-sample --type='json' \
  -p='[{"op": "replace", "path": "/spec/replicas", "value": 5}]'
```

The controller will update the Deployment automatically.

## Delete an app

```bash
kubectl delete appdeployment webapp-sample
```

The controller cleans up the Deployment and Service via finalizers.

## Webhook TLS

Admission webhooks require TLS. The manifests in `config/webhook/` include
cert-manager annotations for automatic certificate provisioning:

```bash
# Install cert-manager (if not already installed)
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/latest/download/cert-manager.yaml

# The webhook configuration will automatically request a certificate
kubectl apply -k config/manager/
```

For development without cert-manager, set `ENABLE_WEBHOOKS=false`:

```bash
helm install k8s-operator-go ./deploy/chart --set webhooks.enabled=false
# or patch the deployment:
kubectl set env deployment/k8s-operator-go -n k8s-operator-system ENABLE_WEBHOOKS=false
```

## Metrics

If Prometheus is running, scrape the operator's metrics endpoint:

```yaml
apiVersion: monitoring.coreos.com/v1
kind: ServiceMonitor
metadata:
  name: k8s-operator-go
  namespace: k8s-operator-system
spec:
  selector:
    matchLabels:
      app.kubernetes.io/name: k8s-operator-go
  endpoints:
    - port: metrics
      path: /metrics
```
