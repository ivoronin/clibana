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
clibana -u https://logs.internal -i "pods-*" search -s now-2h -e now-1h "pod_name:*nginx*"
clibana -u https://logs.internal -i "pods-*" mappings
clibana -u https://logs.internal -i "pods-*" indices
```

Most options can be set using environment variables. Check `clibana -h` for additional details.

### AWS Support

1. Clibana supports the `aws://` scheme to specify an AWS Managed OpenSearch Domain name, which will automatically resolve to its endpoint:
   ```bash
   clibana -u aws://logs-internal -i "logs-*" search "error"
   ```

2. When using the `aws://` scheme, AWS authentication is enabled by default. You can override this by providing username and password file for basic authentication:
   ```bash
   echo 'mypassword' > ~/.clibana-password
   chmod 600 ~/.clibana-password
   clibana -u aws://logs-internal -i "logs-*" -U user --password-file ~/.clibana-password search "error"
   ```

### Authentication

For basic authentication, use a password file (recommended for security):

```bash
# Create password file
echo 'your-password' > ~/.clibana-password
chmod 600 ~/.clibana-password

# Use with clibana (always specify -u and -i)
clibana -u https://logs.internal -i "logs-*" -U admin --password-file ~/.clibana-password search "error"
```

### Environment Variables

Environment variables can be used to avoid repeating options:

```bash
export CLIBANA_URL="https://logs.internal"
export CLIBANA_INDEX="pods-*"

# Now you can omit -u and -i flags
clibana search "error"
clibana search -f "level:ERROR"
```
