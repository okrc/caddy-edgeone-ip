# trusted_proxy Module for Caddy

This module fetches EdgeOne IP addresses using their API.

# Example config

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

# Default Values

## Configuration Options

| Name         | Description                                                       | Type     | Default     |
|--------------|-------------------------------------------------------------------|----------|-------------|
| `zone_id`    | The ID of your EdgeOne zone to retrieve IP addresses for         | string   | required    |
| `secret_id`  | Tencent Cloud API Secret ID used for authentication              | string   | required    |
| `secret_key` | Tencent Cloud API Secret Key used for authentication             | string   | required    |
| `interval`   | How often to fetch the latest EdgeOne IP ranges                  | duration | `1h`        |
| `timeout`    | Maximum time to wait for a response from the EdgeOne API         | duration | no timeout  |
