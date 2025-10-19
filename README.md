# Hetzner DDNS Controller
A Kubernetes controller to automatically update Hetzner DNS records with your current public IP address. Perfect for home labs or self-hosted services with dynamic IP addresses.

## Features
Automatic IP detection and DNS record updates
Configurable refresh intervals
Multiple DNS records supported
Kubernetes native deployment

## How It Works
The controller periodically checks your public IP address from redundant sources (checkip.amazonaws.com, api.ipify.org, icanhazip.com) and updates configured DNS records in Hetzner DNS if changes are detected. This ensures your domain always points to your current IP address, even when it changes.

## Prerequisites
Kubernetes cluster
Hetzner DNS account
GitHub account (for container registry access)

## Installation
### 1. Create Namespace
```
kubectl create namespace hetzner-ddns
```

### 2. Container Registry Access
Create a GitHub Personal Access Token (PAT) for pulling the container image into your cluster:
- Go to GitHub → Settings → Developer Settings → Personal Access Tokens → Fine-grained tokens
- Set Repository Access to "Only select repositories"
- Select the hetzner-ddns repository
- Under "Repository permissions", set Contents to "Read-only"
- Create the registry secret:
```
kubectl create secret docker-registry ghcr-secret \
--namespace hetzner-ddns \
--docker-server=ghcr.io \
--docker-username=<your-github-username> \
--docker-password=<your-github-pat>
```

### 3. Hetzner API Token
Granting the container access to your Hetzner DNS API
- Log into your Hetzner DNS Console
- Click the top right corner to expand the menu
- Select "API Tokens"
- Click "Create API Token"
- Create the Kubernetes secret:
```
kubectl create secret generic hetzner-ddns-secret \
  --namespace hetzner-ddns \
  --from-literal=HETZNER_API_TOKEN=<your-hetzner-token>
```
- Or via YAML with the provided k8s/secret.yam
```
kubectl apply -f k8s/secret.yaml
```
### 4. Edit the configMap in k8s/configmap.yaml and enter your DNS records
See the section Configuration Options for more details

### 5. Deploy the Controller
```
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/deployment.yaml
```
### 6. Verify Installation
```
kubectl -n hetzner-ddns get pods
kubectl -n hetzner-ddns logs -f deployment/hetzner-ddns-deployment
```

## Configuration Options
### Zone ID
To find your Zone ID:
- Go to Hetzner DNS Console
- Click on your domain
- The Zone ID is shown in the overview
  
### Refresh Interval
Set refreshInterval in the ConfigMap to control how often the controller checks for IP changes (in minutes).

### DNS Records
Configure multiple records under the same zone:

name: Subdomain name
type: Record type (typically "A" for IPv4)
ttl: Time to live in seconds


## Troubleshooting
Check controller logs:

## Contributing
Pull requests are welcome!

## License
This project is licensed under the MIT License - see the LICENSE file for details.
