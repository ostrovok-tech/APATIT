# APATIT Helm Chart

A Helm chart for deploying APATIT, a set of Prometheus exporters for [ping-admin.com](https://ping-admin.com/).

## Description

APATIT is a Prometheus exporter that collects metrics from ping-admin.com monitoring service. This Helm chart provides a complete Kubernetes deployment configuration with security best practices, including non-root execution, read-only filesystem, and minimal capabilities.

## Prerequisites

- Kubernetes 1.19+
- Helm 3.0+
- Access to the container image: `ghcr.io/ostrovok-tech/apatit`

## Installation

### Add the chart repository (if applicable)

```bash
helm repo add <repo-name> <repo-url>
helm repo update
```

### Install the chart

```bash
helm install my-apatit ./
```

Or with a custom values file:

```bash
helm install my-apatit ./ -f my-values.yaml
```

## Configuration

The following table lists the configurable parameters and their default values.

### General Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `replicaCount` | Number of replicas | `1` |
| `image.repository` | Container image repository | `ghcr.io/ostrovok-tech/apatit` |
| `image.tag` | Container image tag (defaults to chart appVersion) | `""` |
| `image.pullPolicy` | Image pull policy | `IfNotPresent` |
| `imagePullSecrets` | Secrets for pulling images from private registries | `[]` |
| `nameOverride` | Override the chart name | `""` |
| `fullnameOverride` | Override the full name | `""` |

### Security Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `securityContext.runAsUser` | User ID to run the container | `65534` |
| `securityContext.runAsNonRoot` | Run as non-root user | `true` |
| `securityContext.readOnlyRootFilesystem` | Mount root filesystem as read-only | `true` |
| `securityContext.allowPrivilegeEscalation` | Allow privilege escalation | `false` |
| `securityContext.privileged` | Run in privileged mode | `false` |
| `securityContext.capabilities.drop` | Capabilities to drop | `["ALL"]` |

### Service Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `service.type` | Kubernetes service type | `ClusterIP` |
| `service.port` | Service port | `8080` |

### Ingress Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `ingress.enabled` | Enable ingress | `false` |
| `ingress.className` | Ingress class name | `""` |
| `ingress.annotations` | Ingress annotations | `{}` |
| `ingress.hosts` | Ingress hosts configuration | See values.yaml |
| `ingress.tls` | TLS configuration | `[]` |

### Resource Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `resources.requests.cpu` | CPU request | `15m` |
| `resources.requests.memory` | Memory request | `64Mi` |
| `resources.limits.cpu` | CPU limit | `100m` |
| `resources.limits.memory` | Memory limit | `128Mi` |

### Health Check Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `livenessProbe` | Liveness probe configuration | HTTP GET on `/metrics` |
| `readinessProbe` | Readiness probe configuration | HTTP GET on `/metrics` |

### Application Configuration

| Parameter | Description | Default |
|-----------|-------------|---------|
| `env` | Environment variables | `{}` |
| `envSecrets` | Environment variables from secrets | `[]` |
| `args` | Command-line arguments | `[]` |

### Additional Parameters

| Parameter | Description | Default |
|-----------|-------------|---------|
| `podAnnotations` | Pod annotations | `{}` |
| `podLabels` | Pod labels | See values.yaml |
| `volumes` | Additional volumes | `[]` |
| `volumeMounts` | Additional volume mounts | `[]` |
| `nodeSelector` | Node selector | `{}` |
| `tolerations` | Pod tolerations | `[]` |
| `affinity` | Pod affinity | `{}` |

## Application Configuration

APATIT requires configuration to connect to ping-admin.com API. You can configure it using environment variables or command-line arguments.

### Required Configuration

- **API_KEY**: API key from [ping-admin.com/users/edit/](https://ping-admin.com/users/edit/)
- **TASK_IDS**: Task IDs from the ID column in [ping-admin.com/tasks/](https://ping-admin.com/tasks/)

### Configuration Methods

#### Method 1: Environment Variables

```yaml
env:
  API_KEY: "<your_api_key>"
  TASK_IDS: "<task_ids>"
  LISTEN_ADDRESS: ":8080"
  LOG_LEVEL: "info"
```

#### Method 2: Environment Variables from Secrets (Recommended)

```yaml
envSecrets:
  - name: API_KEY
    secretName: ping-admin-secret
    secretKey: readonly-api-key
  - name: TASK_IDS
    secretName: ping-admin-secret
    secretKey: task-ids
```

#### Method 3: Command-line Arguments

```yaml
args:
  - "--api-key=<your_api_key>"
  - "--task-ids=<task_ids>"
  - "--listen-address=:8080"
  - "--log-level=info"
```

### Available Configuration Options

| Environment Variable | Command-line Argument | Description |
|---------------------|----------------------|-------------|
| `API_KEY` | `--api-key` | API key from ping-admin.com |
| `TASK_IDS` | `--task-ids` | Task IDs to monitor |
| `LISTEN_ADDRESS` | `--listen-address` | Address to listen on | `:8080` |
| `LOG_LEVEL` | `--log-level` | Logging level | `info` |
| `LOCATIONS_FILE` | `--locations-file` | Locations file path | `locations.json` |
| `ENG_MP_NAMES` | `--eng-mp-names` | Use English MP names | `true` |
| `REFRESH_INTERVAL` | `--refresh-interval` | Refresh interval | `3m` |
| `API_UPDATE_DELAY` | `--api-update-delay` | API update delay | `4m` |
| `API_DATA_TIME_STEP` | `--api-data-time-step` | API data time step | `3m` |
| `MAX_ALLOWED_STALENESS_STEPS` | `--max-allowed-staleness-steps` | Max allowed staleness steps | `3` |
| `MAX_REQUESTS_PER_SECOND` | `--max-requests-per-second` | Max requests per second | `2` |
| `REQUEST_DELAY` | `--request-delay` | Request delay | `2s` |
| `REQUEST_RETRIES` | `--request-retries` | Request retries | `3` |

## Examples

### Basic Installation

```bash
helm install apatit ./
```

### Installation with Custom Values

Create a `custom-values.yaml`:

```yaml
replicaCount: 2

env:
  API_KEY: "your-api-key-here"
  TASK_IDS: "1,2,3"
  LOG_LEVEL: "debug"

service:
  type: ClusterIP
  port: 8080

ingress:
  enabled: true
  className: "nginx"
  annotations:
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
  hosts:
    - host: apatit.example.com
      paths:
        - path: /
          pathType: Prefix
  tls:
    - secretName: apatit-tls
      hosts:
        - apatit.example.com

resources:
  requests:
    cpu: 50m
    memory: 128Mi
  limits:
    cpu: 200m
    memory: 256Mi
```

Install with custom values:

```bash
helm install apatit ./ -f custom-values.yaml
```

### Using Secrets for Sensitive Data

First, create a Kubernetes secret:

```bash
kubectl create secret generic ping-admin-secret \
  --from-literal=readonly-api-key='your-api-key' \
  --from-literal=task-ids='1,2,3'
```

Then configure the chart to use the secret:

```yaml
envSecrets:
  - name: API_KEY
    secretName: ping-admin-secret
    secretKey: readonly-api-key
  - name: TASK_IDS
    secretName: ping-admin-secret
    secretKey: task-ids
```

### Prometheus Scraping Configuration

To enable Prometheus scraping, add annotations:

```yaml
podAnnotations:
  prometheus.io/scrape: "true"
  prometheus.io/port: "8080"
  prometheus.io/path: "/metrics"
```

## Metrics

The exporter exposes Prometheus metrics at the `/metrics` endpoint. The default port is `8080`.

## Upgrading

```bash
helm upgrade apatit ./
```

Or with custom values:

```bash
helm upgrade apatit ./ -f custom-values.yaml
```

## Uninstalling

```bash
helm uninstall apatit
```

## Troubleshooting

### Check Pod Status

```bash
kubectl get pods -l app.kubernetes.io/name=apatit
```

### View Pod Logs

```bash
kubectl logs -l app.kubernetes.io/name=apatit
```

### Check Service

```bash
kubectl get svc -l app.kubernetes.io/name=apatit
```

### Test Metrics Endpoint

```bash
kubectl port-forward svc/apatit 8080:8080
curl http://localhost:8080/metrics
```

## Security Considerations

This chart is configured with security best practices:

- Runs as non-root user (UID 65534)
- Read-only root filesystem
- No privilege escalation
- All capabilities dropped
- Minimal resource requests/limits

## Links

- [Source Code](https://github.com/ostrovok-tech/apatit)
- [ping-admin.com](https://ping-admin.com/)
- [Helm Documentation](https://helm.sh/docs/)

## Maintainers

- **Ostrovok! Tech** - [GitHub](https://github.com/ostrovok-tech)

## License

Please refer to the source repository for license information.

