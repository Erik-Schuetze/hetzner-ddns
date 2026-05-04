# Hetzner DDNS Controller
A Kubernetes controller to automatically update migrated Hetzner DNS records with your current public IP address using the Hetzner Cloud DNS API. Perfect for home labs or self-hosted services with dynamic IP addresses.

## Features
Automatic IP detection and DNS record updates
Configurable refresh intervals
Multiple DNS records supported
Kubernetes native deployment

## How It Works
The controller periodically checks your public IP address from redundant sources (checkip.amazonaws.com, api.ipify.org, icanhazip.com) and updates configured DNS records in Hetzner DNS if changes are detected. If a record is declared in config but missing in Hetzner, the controller creates it automatically. This ensures your domain always points to your current IP address, even when it changes.

> [!IMPORTANT]
> This version targets Hetzner's new Cloud DNS API and requires zones to be migrated to Hetzner Console first. The legacy `dns.hetzner.com` API is no longer used by this controller.

## Prerequisites
Kubernetes cluster
Hetzner Console project with migrated DNS zones
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

### 3. Hetzner Cloud API Token
Grant the container access to your migrated DNS zones:
- Migrate each zone to Hetzner Console before deploying this version.
- Create a Hetzner Cloud API token in Hetzner Console with access to the project that contains the migrated zones.
- Create the Kubernetes secret:
```
kubectl create secret generic hetzner-ddns-secret \
  --namespace hetzner-ddns \
  --from-literal=HETZNER_CLOUD_API_TOKEN=<your-hetzner-cloud-token>
```
- Or via YAML with the provided `k8s/secret.yaml`
```
kubectl apply -f k8s/secret.yaml
```
### 4. Edit the ConfigMap in k8s/configmap.yaml and enter your DNS zones and records
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
### Zone Name
Set `zone_name` to the migrated zone's domain name, for example `example.com`.
  
### Refresh Interval
Set `refresh_interval` in the ConfigMap to control how often the controller checks for IP changes (in minutes).

### DNS Records
Configure multiple records under the same zone:

name: Subdomain name
type: Record type (typically "A" for IPv4)
ttl: Time to live in seconds

### Migration checklist
Before rolling out this version:
1. Verify the zone satisfies Hetzner's migration requirements.
2. Migrate the zone to Hetzner Console.
3. Create a new Hetzner Cloud API token in Hetzner Console.
4. Update the Kubernetes secret and deploy the new image.

See `docs/migration-guide.md` for the release-specific migration steps.


## Troubleshooting
Check controller logs:

## Contributing
Pull requests are welcome!

## License
This project is licensed under the MIT License - see the LICENSE file for details.
