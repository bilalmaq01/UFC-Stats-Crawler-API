# ---------------------------------------------------------------------------
# ECS task roles
# ---------------------------------------------------------------------------

data "aws_iam_policy_document" "ecs_assume" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["ecs-tasks.amazonaws.com"]
    }
  }
}

# Execution role: pulls the image, reads SSM secrets, writes logs.
resource "aws_iam_role" "execution" {
  name               = "${var.project}-execution"
  assume_role_policy = data.aws_iam_policy_document.ecs_assume.json
  tags               = local.tags
}

resource "aws_iam_role_policy_attachment" "execution_managed" {
  role       = aws_iam_role.execution.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy"
}

data "aws_iam_policy_document" "execution_secrets" {
  statement {
    sid       = "ReadSsmParameters"
    actions   = ["ssm:GetParameters"]
    resources = [aws_ssm_parameter.database_url.arn]
  }
  statement {
    sid       = "DecryptSsmDefaultKey"
    actions   = ["kms:Decrypt"]
    resources = ["*"] # default aws/ssm key; scope to a CMK ARN if you use one
  }
}

resource "aws_iam_role_policy" "execution_secrets" {
  name   = "ssm-read"
  role   = aws_iam_role.execution.id
  policy = data.aws_iam_policy_document.execution_secrets.json
}

# Task role: what the running crawler itself can do. It only talks to an external
# DB and website, so no AWS permissions are needed — kept for clarity/future use.
resource "aws_iam_role" "task" {
  name               = "${var.project}-task"
  assume_role_policy = data.aws_iam_policy_document.ecs_assume.json
  tags               = local.tags
}

# ---------------------------------------------------------------------------
# EventBridge Scheduler role: allowed to RunTask and pass the task roles.
# ---------------------------------------------------------------------------

data "aws_iam_policy_document" "scheduler_assume" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["scheduler.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "scheduler" {
  name               = "${var.project}-scheduler"
  assume_role_policy = data.aws_iam_policy_document.scheduler_assume.json
  tags               = local.tags
}

data "aws_iam_policy_document" "scheduler" {
  statement {
    sid       = "RunCrawlerTask"
    actions   = ["ecs:RunTask"]
    resources = ["${aws_ecs_task_definition.crawler.arn_without_revision}:*"]

    condition {
      test     = "ArnLike"
      variable = "ecs:cluster"
      values   = [aws_ecs_cluster.this.arn]
    }
  }
  statement {
    sid       = "PassTaskRoles"
    actions   = ["iam:PassRole"]
    resources = [aws_iam_role.execution.arn, aws_iam_role.task.arn]
  }
}

resource "aws_iam_role_policy" "scheduler" {
  name   = "run-task"
  role   = aws_iam_role.scheduler.id
  policy = data.aws_iam_policy_document.scheduler.json
}
