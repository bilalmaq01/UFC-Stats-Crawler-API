terraform {
  required_version = ">= 1.6"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.60"
    }
  }

  # First tofu for this project → local state. When you're ready to share state
  # across machines/CI, uncomment and point at an S3 bucket you create first.
  # backend "s3" {
  #   bucket = "your-tofu-state-bucket"
  #   key    = "ufc-crawler/terraform.tfstate"
  #   region = "us-east-1"
  # }
}

provider "aws" {
  region = var.aws_region
}
