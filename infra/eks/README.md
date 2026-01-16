# EKS Cluster

This cluster is created using `eksctl` on top of an existing VPC
provisioned via Terraform (`infra/vpc`).

## Create
```bash
eksctl create cluster -f cluster.yaml

```

## Delete

```bash
eksctl delete cluster -f cluster.yaml

```
