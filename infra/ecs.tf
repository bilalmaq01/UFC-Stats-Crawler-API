resource "aws_ecs_cluster" "this" {
  name = var.project
  tags = local.tags
}

resource "aws_cloudwatch_log_group" "crawler" {
  name              = "/ecs/${var.project}"
  retention_in_days = var.log_retention_days
  tags              = local.tags
}

resource "aws_ecs_task_definition" "crawler" {
  family                   = var.project
  requires_compatibilities = ["FARGATE"]
  network_mode             = "awsvpc"
  cpu                      = var.task_cpu
  memory                   = var.task_memory
  execution_role_arn       = aws_iam_role.execution.arn
  task_role_arn            = aws_iam_role.task.arn

  runtime_platform {
    operating_system_family = "LINUX"
    cpu_architecture        = "ARM64"
  }

  container_definitions = jsonencode([
    {
      name      = "crawler"
      image     = "${aws_ecr_repository.crawler.repository_url}:${var.image_tag}"
      essential = true
      command   = var.crawler_command

      secrets = [
        { name = "DATABASE_URL", valueFrom = aws_ssm_parameter.database_url.arn },
      ]

      logConfiguration = {
        logDriver = "awslogs"
        options = {
          "awslogs-group"         = aws_cloudwatch_log_group.crawler.name
          "awslogs-region"        = var.aws_region
          "awslogs-stream-prefix" = "crawler"
        }
      }
    }
  ])

  tags = local.tags
}
