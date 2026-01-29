#####################################
# Data
#####################################
data "aws_availability_zones" "available" {
  state = "available"
}

#####################################
# Locals
#####################################
locals {
  azs = slice(data.aws_availability_zones.available.names, 0, var.az_count)

  # /24 public + /24 private per AZ
  public_subnets  = [for i, az in local.azs : cidrsubnet(var.vpc_cidr, 8, i)]
  private_subnets = [for i, az in local.azs : cidrsubnet(var.vpc_cidr, 8, i + 10)]
}

#####################################
# VPC
#####################################
module "vpc" {
  source  = "terraform-aws-modules/vpc/aws"
  version = "~> 5.0"

  name = "${var.name}-vpc"
  cidr = var.vpc_cidr

  azs             = local.azs
  public_subnets  = local.public_subnets
  private_subnets = local.private_subnets

  enable_dns_support   = true
  enable_dns_hostnames = true

  enable_nat_gateway = true
  single_nat_gateway  = true

  # Required later for EKS / ALB
  public_subnet_tags = {
    "kubernetes.io/role/elb" = "1"
  }

  private_subnet_tags = {
    "kubernetes.io/role/internal-elb" = "1"
  }

  tags = merge(
    {
      Project = "go-microservices"
      Owner   = "tirdad"
    },
    var.tags
  )
}

#####################################
# Redis Security Group
#####################################
resource "aws_security_group" "redis" {
  name        = "${var.name}-redis-sg"
  description = "Allow Redis access from within VPC"
  vpc_id      = module.vpc.vpc_id

  ingress {
    description = "Redis from VPC"
    from_port   = 6379
    to_port     = 6379
    protocol    = "tcp"
    cidr_blocks = [var.vpc_cidr]
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }

  tags = {
    Name = "${var.name}-redis-sg"
  }
}

#####################################
# ElastiCache Serverless (Valkey / Redis)
#####################################
resource "aws_elasticache_serverless_cache" "redis" {
  name   = "${var.name}-elasticache"
  engine = "valkey" # AWS default for serverless (Redis-compatible)

  major_engine_version = "8"

  subnet_ids         = module.vpc.private_subnets
  security_group_ids = [aws_security_group.redis.id]

  cache_usage_limits {
    data_storage {
      maximum = 1
      unit    = "GB"
    }

    ecpu_per_second {
      maximum = 1000
    }
  }

  description = "Serverless Redis for go-microservices"

  tags = {
    Name = "${var.name}-elasticache"
  }
}
