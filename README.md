# trusted_proxy Module for Caddy

This module fetches EdgeOne IP addresses using their API.

## Example config

Add the following configuration under the appropriate server block in the global options:

```
trusted_proxies edgeone {
    zone_id xxx
    secret_id xxx
    secret_key xxx
    interval 168h
    timeout 15s
}

```

You can also use the [caddy-combine-ip-ranges](https://github.com/fvbommel/caddy-combine-ip-ranges) module to combine multiple IP sources:
```
trusted_proxies combine {
    static private_ranges
    cloudflare {
        interval 72h
        timeout 5s
    }
    edgeone {
        zone_id xxx
        secret_id xxx
        secret_key xxx
        interval 168h
        timeout 5s
    }
}
```

## Configuration Options

> `zone_id`, `secret_id`, and `secret_key` are only required when **Origin Protection** is enabled in your EdgeOne configuration.

| Name         | Description                                                                 | Type     | Default     |
|--------------|-----------------------------------------------------------------------------|----------|-------------|
| `zone_id`    | EdgeOne zone ID (required only if Origin Protection is enabled)             | string   | none        |
| `secret_id`  | Tencent Cloud API Secret ID (used only if Origin Protection is enabled)     | string   | none        |
| `secret_key` | Tencent Cloud API Secret Key (used only if Origin Protection is enabled)    | string   | none        |
| `area`       | Geographic region: `global`, `mainland-china`, or `overseas`                | string   | all regions |
| `version`    | IP version: `v4`, `v6`, or empty for both                                   | string   | both        |
| `interval`   | How often to refresh EdgeOne IP ranges                                      | duration | `1h`        |
| `timeout`    | Max time to wait for a response from the EdgeOne API                        | duration | no timeout  |
