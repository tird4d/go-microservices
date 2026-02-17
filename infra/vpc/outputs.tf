output "vpc_id" {
  value = module.vpc.vpc_id
}

output "public_subnet_ids" {
  value = module.vpc.public_subnets
}

output "private_subnet_ids" {
  value = module.vpc.private_subnets
}

output "redis_endpoint" {
  value = aws_elasticache_serverless_cache.redis.endpoint[0].address
}

output "redis_port" {
  value = aws_elasticache_serverless_cache.redis.endpoint[0].port
}

output "redis_sg_id" {
  value = aws_security_group.redis.id
}
