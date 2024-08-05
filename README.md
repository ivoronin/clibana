# Clibana
![GitHub release (with filter)](https://img.shields.io/github/v/release/ivoronin/clibana)
[![Go Report Card](https://goreportcard.com/badge/github.com/ivoronin/clibana)](https://goreportcard.com/report/github.com/ivoronin/clibana)
![GitHub last commit (branch)](https://img.shields.io/github/last-commit/ivoronin/clibana/main)
![GitHub Workflow Status (with event)](https://img.shields.io/github/actions/workflow/status/ivoronin/clibana/main.yml)
![GitHub top language](https://img.shields.io/github/languages/top/ivoronin/clibana)

## Description

Clibana is a command-line interface (CLI) tool for OpenSearch that offers Kibana-like log searching and live tailing (`-f` support) capabilities.

## Features

- **Log Search**: Execute searches on multiple OpenSearch indices using Lucene query syntax. Clibana can output specified fields or export data in NDJSON format.
- **Live Tailing**: Monitor logs in real-time using the `-f` option, similar to `tail -f`.
- **Cluster Exploration**: List index information and field mappings.
- **AWS and Basic Authentication**: Supports both AWS SigV4 and Basic Authentication methods.

## Examples

```bash
clibana -H https://logs.internal -i "k8s.containers.*" search -s now-2h -e now-1h "kubernetes.pod_name:*nginx*"
clibana -H https://logs.internal -i "k8s.containers.*" mappings
clibana -H https://logs.internal -i "k8s.containers.*" indices
```

Most options can be set using environment variables. Check `clibana` -h for additional details.

### AWS Support

1. Clibana supports the `aws://` scheme to specify an AWS Managed OpenSearch Domain name as a host, which will automatically resolve to its endpoint. Example: `clibana --host aws://logs-internal`.
2. You can use your AWS credentials to authenticate to an AWS Managed OpenSearch Domain. Set the authentication type to `aws`: `clibana -a aws`.