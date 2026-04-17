# EKS Cluster Management

## Create Cluster

```bash
cd infra/eks

# Create cluster from configuration
eksctl create cluster -f cluster.yaml

# Takes ~15-20 minutes to complete
```

## Delete Cluster

```bash
# WARNING: This deletes everything (nodes, data, services, volumes)

cd infra/eks

# Delete cluster
eksctl delete cluster -f cluster.yaml

# Verify deletion
eksctl get clusters --region eu-central-1
```