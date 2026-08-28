# Kubernetes Deployment

Deploy CP'S Enterprise DCS to Kubernetes.

## Prerequisites

- Kubernetes 1.24+
- kubectl configured
- Docker images built and available: `regional-agent:latest`, `local-agent:latest`

## Deploy

```bash
kubectl apply -f 00-namespace-config.yaml
kubectl apply -f 01-postgres.yaml
kubectl apply -f 02-redis.yaml
kubectl apply -f 03-regional-agent.yaml
kubectl apply -f 04-local-agent.yaml
```

## Verify

```bash
kubectl get pods -n dcs
kubectl get svc -n dcs
kubectl port-forward svc/regional-agent 8080:8080 -n dcs
curl http://localhost:8080/health
```

## Notes

- Update secrets in `00-namespace-config.yaml` before applying.
- For production, replace `emptyDir` volumes with `PersistentVolumeClaim`.
- Consider using `StatefulSet` for PostgreSQL with proper PVCs.
