Kubernetes Volume Backup to S3 (Go)

This repo contains a Helm chart and container image that back up arbitrary directories from containers in Kubernetes pods to S3. The backup logic is implemented in Go and shells out to `kubectl` and `aws` for maximum compatibility without external Go dependencies.

How it works
- Resolves a pod by Deployment/StatefulSet/Pod name or by label selector.
- Streams `tar.gz` directly from the container to S3 via a `kubectl exec | aws s3 cp -` pipeline (без локальных файлов).
- Optionally prunes backups by TTL (`BACKUP_TTL`), deleting objects older than the specified duration.

Environment variables
- `KUBE_NAMESPACE`/`NAMESPACE`: Target namespace (default: `default`).
- `WORKLOAD_KIND`: `deployment` | `statefulset` | `pod` (optional if using selector).
- `WORKLOAD_NAME`: Name of workload (required if `WORKLOAD_KIND` is set).
- `POD_LABEL_SELECTOR`: Direct label selector, e.g. `app=foo,role=bar` (alternative to kind/name).
- `CONTAINER`: Container name in the pod (optional; авто‑детект если контейнер один).
- `SRC_PATH`: REQUIRED. Path inside the container to archive.
- `BACKUP_NAME_PREFIX`: Optional filename prefix; falls back to workload/pod name.
- `AWS_ACCESS_KEY_ID` / `AWS_SECRET_ACCESS_KEY` / `AWS_DEFAULT_REGION`: AWS creds and region.
- `AWS_S3_ENDPOINT_URL`: Optional S3-compatible endpoint (e.g., MinIO).
- `AWS_BUCKET_NAME`: Target bucket (required).
- `AWS_BUCKET_BACKUP_PATH`: Prefix/folder in the bucket (optional; e.g., `backups/app1`).
- `BACKUP_TTL`: TTL for backups, e.g. `7d` or `168h`. All objects older than TTL are deleted each run.

Build locally
1. Ensure Docker is available.
2. Build the image:
   docker build -t yourrepo/k8s-s3-backup:latest .
3. Push/tag as needed, or use with the included Helm chart.

Helm usage
- Chart under `helm-chart/` (name: k8s-s3-backup) wires env vars for the CronJob using generic inputs.
- Values key for the job: `cronjobs.backup`.

Notes
- Requires `kubectl` and `aws` CLIs in the runtime image; the Dockerfile installs both.
- The Go binary avoids extra dependencies; it uses the CLIs to interact with Kubernetes and S3.

Repository layout
- `cmd/backup`: CLI entrypoint.
- `internal/config`: config from environment.
- `internal/kube`: kubectl helpers (pod/container resolve).
- `internal/s3`: S3 helpers (URLs, TTL pruning).
- `internal/backup`: backup orchestration and streaming.
