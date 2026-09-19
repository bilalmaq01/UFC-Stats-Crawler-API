# UFC crawler — Fargate deployment (OpenTofu)

Runs `cmd/crawler` as a scheduled AWS Fargate task: EventBridge Scheduler fires a
one-shot ECS task on a cron, the task pulls the image from ECR, reads secrets from
SSM Parameter Store, crawls, writes to your external Postgres, and exits.

```
EventBridge Scheduler ──RunTask──▶ ECS Fargate task ──▶ ECR (image)
   (cron)                             │                  SSM (secrets)
                                      ▼                  CloudWatch (logs)
                              external Postgres
```

Everything is ARM64/Graviton (matches the Dockerfile). No VPC/NAT is created — the
task uses the default VPC's public subnets with a public IP for outbound access.

## Prerequisites

- OpenTofu, Docker (with buildx), and the AWS CLI, all authenticated to your account.
- The database reachable from the public internet (Fargate connects over egress).

## One-time setup

```sh
cd infra
cp terraform.tfvars.example terraform.tfvars   # fill in database_url, cookie, region
tofu init
```

## Deploy

The image must exist in ECR before a task can pull it, so create the repo first,
push, then apply the rest.

```sh
# 1. Create just the ECR repo
tofu apply -target=aws_ecr_repository.crawler

# 2. Log in and push the image (run from the repo root)
cd ..
REGION=$(cd infra && tofu output -raw ecr_repository_url | cut -d. -f4)
REPO=$(cd infra && tofu output -raw ecr_repository_url)
aws ecr get-login-password --region "$REGION" | docker login --username AWS --password-stdin "${REPO%/*}"

docker buildx build --platform linux/arm64 -t "$REPO:latest" --push .

# 3. Apply everything else
cd infra && tofu apply
```

## Test before trusting the schedule

Kick off the task manually and watch the logs:

```sh
CLUSTER=$(tofu output -raw cluster_name)
TASKDEF=$(tofu output -raw task_definition_family)
SUBNET=$(tofu output -json subnet_ids | jq -r '.[0]')
SG=$(tofu output -raw security_group_id)

aws ecs run-task \
  --cluster "$CLUSTER" \
  --task-definition "$TASKDEF" \
  --launch-type FARGATE \
  --network-configuration "awsvpcConfiguration={subnets=[$SUBNET],securityGroups=[$SG],assignPublicIp=ENABLED}"

# Tail logs
aws logs tail "$(tofu output -raw log_group)" --follow
```

You should see the same `event:` / `fight ... -> N stat rows` / `crawl complete`
output you saw locally.

## Redeploying new code

```sh
docker buildx build --platform linux/arm64 -t "$REPO:latest" --push .
# force the scheduler onto the new image
aws ecs run-task ...   # or just wait for the next scheduled run
```

Because the task pulls `:latest`, the next scheduled run picks up the new image.
For reproducible deploys, push an immutable tag (e.g. git SHA) and set `image_tag`.

## Tuning

- **Schedule**: `schedule_expression` (e.g. `rate(6 hours)`, `cron(0 12 * * ? *)`).
- **What to crawl**: `crawler_command` — `["all"]` full refresh vs `["all","5"]` recent.
- **Pause auto-runs**: `schedule_enabled = false`.
- **Crawler politeness/speed** lives in the Go code (`newCollector`: `Parallelism`,
  `RandomDelay`). Bump the task `task_cpu`/`task_memory` if a full crawl needs it.

## Teardown

```sh
tofu destroy
```
