# Migration guide: legacy Hetzner DNS API -> Hetzner Cloud DNS API

This release switches `hetzner-ddns` from Hetzner's legacy DNS API at `dns.hetzner.com` to the new Hetzner Cloud DNS API used by Hetzner Console.

## What changed

This release includes these breaking changes:

- The controller no longer uses the legacy `dns.hetzner.com/api/v1` API.
- Zones must already be migrated to Hetzner Console before this version can manage them.
- Configuration now uses `zone_name` instead of `zone_id`.
- The Kubernetes secret/env var now uses `HETZNER_CLOUD_API_TOKEN` instead of `HETZNER_API_TOKEN`.

## Before you upgrade

Complete these manual steps first:

1. Verify each zone satisfies Hetzner's migration requirements.
2. Migrate each zone to Hetzner Console.
3. Create a Hetzner Cloud API token in Hetzner Console for the project that contains the migrated zones.

## Config changes

Update your ConfigMap from:

```yaml
hetzner:
  zones:
    - zone_id: "example.com"
      records:
        - name: "@"
          type: "A"
          ttl: 3600
```

to:

```yaml
hetzner:
  zones:
    - zone_name: "example.com"
      records:
        - name: "@"
          type: "A"
          ttl: 3600
```

`zone_name` should be the migrated DNS zone name, for example `example.com`.

## Secret and deployment changes

Update your secret from:

```yaml
stringData:
  HETZNER_API_TOKEN: "<old-token>"
```

to:

```yaml
stringData:
  HETZNER_CLOUD_API_TOKEN: "<new-cloud-token>"
```

Update the Deployment env var reference from `HETZNER_API_TOKEN` to `HETZNER_CLOUD_API_TOKEN`.

## Recommended rollout order

1. Migrate the DNS zones in Hetzner.
2. Create the new Hetzner Cloud API token.
3. Update your Kubernetes manifests or generated secrets/config.
4. Deploy the new controller image.
5. Watch the controller logs and confirm the expected rrsets are updated successfully.

## What does not change

- The controller still updates configured records based on the detected public IP.
- `refresh_interval` and the per-record `name`, `type`, and `ttl` fields remain the same.
- The Kubernetes deployment model stays the same apart from the renamed secret/env var.

## If you have not migrated yet

Do not deploy this release until the target zones are migrated to Hetzner Console. If you still depend on the legacy DNS API, stay on the previous release until your zones and token are ready.
