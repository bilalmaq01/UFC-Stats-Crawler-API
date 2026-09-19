# Secrets injected into the task as env vars at runtime (encrypted with the
# default aws/ssm KMS key). Values come from your gitignored terraform.tfvars.

resource "aws_ssm_parameter" "database_url" {
  name  = "/${var.project}/DATABASE_URL"
  type  = "SecureString"
  value = var.database_url
  tags  = local.tags
}
