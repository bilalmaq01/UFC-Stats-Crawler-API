variable "aws_region" {
  description = "AWS region to deploy into."
  type        = string
  default     = "us-east-1"
}

variable "project" {
  description = "Name prefix for all resources (ECR repo, cluster, roles, etc.)."
  type        = string
  default     = "ufc-crawler"
}

variable "image_tag" {
  description = "ECR image tag the task should run."
  type        = string
  default     = "latest"
}

variable "task_cpu" {
  description = "Fargate task CPU units (256 = 0.25 vCPU)."
  type        = string
  default     = "512"
}

variable "task_memory" {
  description = "Fargate task memory in MiB."
  type        = string
  default     = "1024"
}

variable "crawler_command" {
  description = "Container command override, e.g. [\"all\",\"5\"] or [\"all\"] for a full refresh."
  type        = list(string)
  default     = ["all", "5"]
}

variable "schedule_expression" {
  description = "EventBridge Scheduler expression (cron/rate/at)."
  type        = string
  default     = "cron(0 12 * * ? *)" # 12:00 daily
}

variable "schedule_timezone" {
  description = "Timezone for the schedule expression."
  type        = string
  default     = "Etc/UTC"
}

variable "schedule_enabled" {
  description = "Whether the scheduled crawl is active. Set false to deploy but not auto-run."
  type        = bool
  default     = true
}

variable "log_retention_days" {
  description = "CloudWatch log retention for the crawler."
  type        = number
  default     = 30
}

# --- secrets: keep these in a gitignored terraform.tfvars (never commit real values) ---

variable "database_url" {
  description = "Postgres connection string the crawler writes to."
  type        = string
  sensitive   = true
}
