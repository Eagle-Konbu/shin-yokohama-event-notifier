terraform {
  required_version = ">= 1.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = ">= 5.0, < 6.0"
    }
    grafana = {
      source  = "grafana/grafana"
      version = ">= 3.0"
    }
  }
}

provider "aws" {
  region = var.aws_region
}

provider "grafana" {
  url  = var.grafana_url
  auth = var.grafana_auth
}
