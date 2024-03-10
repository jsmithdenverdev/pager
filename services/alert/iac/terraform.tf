terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.38.0"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.6.0"
    }
    archive = {
      source  = "hashicorp/archive"
      version = "~> 2.4.2"
    }
  }

  backend "s3" {
    bucket = "jsmithdenverdev-pager-tf-state"
    key    = "alert-service"
    region = "us-west-1"
  }
}

provider "aws" {
  region = var.aws_region

  # default_tags {
  #   tags = {}
  # }
}