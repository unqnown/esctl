### esctl

esctl is CLI tool for Elasticsearch cluster managing.

### installation

Standard `go install`:

```shell script
go install github.com/unqnown/esctl
```

### configuration

To start using `esctl` immediately with default configuration run:

```shell script
esctl init
```

Default configuration will be added to your `$HOME/.esctl` directory.
You are able to override config location with `$ESCTLCONFIG` env variable.

```yaml
clusters:
  localhost:
    servers:
    - http://localhost:9200
contexts:
  default:
    cluster: localhost
    location: default
context: default
```
