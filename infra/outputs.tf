output "ecr_repository_url" {
  description = "Push the image here."
  value       = aws_ecr_repository.crawler.repository_url
}

output "cluster_name" {
  value = aws_ecs_cluster.this.name
}

output "task_definition_family" {
  value = aws_ecs_task_definition.crawler.family
}

output "log_group" {
  value = aws_cloudwatch_log_group.crawler.name
}

output "subnet_ids" {
  description = "Subnets used for manual `aws ecs run-task` tests."
  value       = data.aws_subnets.default.ids
}

output "security_group_id" {
  value = aws_security_group.crawler.id
}
