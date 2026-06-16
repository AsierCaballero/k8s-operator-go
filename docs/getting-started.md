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
