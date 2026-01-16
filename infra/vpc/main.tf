data "aws_availability_zones" "available" {
  state = "available"
}

locals {
  azs = slice(data.aws_availability_zones.available.names, 0, var.az_count)

  # /24 public + /24 private per AZ (برای 2 AZ میشه 4 تا /24)
  public_subnets  = [for i, az in local.azs : cidrsubnet(var.vpc_cidr, 8, i)]
  private_subnets = [for i, az in local.azs : cidrsubnet(var.vpc_cidr, 8, i + 10)]
}

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
  single_nat_gateway = true

  # Tags مهم برای EKS/LoadBalancer ها (بعداً لازم میشه)
  public_subnet_tags = {
    "kubernetes.io/role/elb" = "1"
  }

  private_subnet_tags = {
    "kubernetes.io/role/internal-elb" = "1"
  }

  tags = merge(
    {
      "Project" = "go-microservices"
      "Owner"   = "tirdad"
    },
    var.tags
  )
}
