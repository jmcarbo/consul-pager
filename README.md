Consul-pager is a consul check monitoring agent.

It currently only supports notification via Slack channels.

Usage
=====

To start the monitoring agent type:

```shell
consul-pager start -config config.yaml
```

The configuration file should look like:

```yaml
channels:
  MySlackChannel:
    channeli:
      subdomain: bbbbb
      token: bbbbbbb
      channel: '#consul'
      user: consul-pager
  MySlackChannel2:
    channeli:
      subdomain: xxxxxxxxx
      token: xxxxxxxxxx
      channel: '#consul'
      user: consul-pager2
defaultmaturity: 30s
```

Channels holds an array of notification channels and defaultmaturity holds the interval that a check must fail before a notification is sent.

You can easily add checks with:

```shell
consul-alert addcheck -name a_check_name -interval 10s -script "ping -c 1 localhost"
```