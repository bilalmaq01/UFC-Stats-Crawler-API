# The crawler only needs outbound internet (ECR, ufcstats.com, external Postgres),
# so it runs in the default VPC's public subnets with a public IP — no NAT needed.

data "aws_vpc" "default" {
  default = true
}

data "aws_subnets" "default" {
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.default.id]
  }
}

resource "aws_security_group" "crawler" {
  name_prefix = "${var.project}-"
  description = "Egress-only security group for the UFC crawler Fargate task"
  vpc_id      = data.aws_vpc.default.id

  egress {
    description = "all outbound"
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = local.tags

  lifecycle {
    create_before_destroy = true
  }
}
