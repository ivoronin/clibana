# clibana

CLI log tailer for OpenSearch with Lucene query syntax and live streaming

[![CI](https://github.com/ivoronin/clibana/actions/workflows/test.yml/badge.svg)](https://github.com/ivoronin/clibana/actions/workflows/test.yml)
[![Release](https://img.shields.io/github/v/release/ivoronin/clibana)](https://github.com/ivoronin/clibana/releases)

[Overview](#overview) · [Features](#features) · [Installation](#installation) · [Usage](#usage) · [Configuration](#configuration) · [Requirements](#requirements) · [License](#license)

```bash
# Search logs from the last 2 hours
clibana -u https://logs.internal -i "pods-*" search -s now-2h "level:ERROR"

# Live tail with field selection
clibana -u https://logs.internal -i "pods-*" search -f -F "@timestamp,message" "pod_name:nginx*"
```

## Overview

Clibana queries OpenSearch indices using Lucene query syntax and streams results to stdout. In follow mode (`-f`), it continuously polls for new documents, similar to `tail -f`. Supports AWS Managed OpenSearch with automatic domain endpoint resolution via the `aws://` scheme and SigV4 authentication.

## Features

- Lucene query syntax with field-specific searches, wildcards, phrases, and boolean operators
- Live tailing mode (`-f`) with adaptive polling (0-5 second intervals based on event rate)
- Field selection for custom output format, or NDJSON for full document output
- AWS Managed OpenSearch support with automatic domain endpoint resolution
- AWS SigV4, HTTP Basic, and cookie-based authentication via dashboard proxy
- Index exploration commands: list indices and field mappings

## Installation

### GitHub Releases

Download from [Releases](https://github.com/ivoronin/clibana/releases).

### Homebrew

```bash
brew install ivoronin/ivoronin/clibana
```

## Usage

### Search

```bash
# Search all logs from the last 5 minutes (default)
clibana -u https://logs.internal -i "logs-*" search "*"

# Search with time range
clibana -u https://logs.internal -i "logs-*" search -s now-2h -e now-1h "level:ERROR"

# Live tail mode
clibana -u https://logs.internal -i "logs-*" search -f "pod_name:web-*"

# Select specific fields (space-separated output)
clibana -u https://logs.internal -i "logs-*" search -F "@timestamp,level,message" "error"
```

### Query Syntax

```bash
*                               # Match all logs
error                           # Search in all fields
level:ERROR                     # Field-specific search
pod_name:nginx*                 # Wildcard search
message:"out of memory"         # Phrase search
level:ERROR AND pod:web-*       # Boolean AND
level:ERROR AND NOT pod:test-*  # Boolean AND NOT
```

### Time Formats

```bash
-s now-5m                    # Last 5 minutes (default)
-s now-2h -e now-1h          # 1-2 hours ago
-s 2024-01-15T10:00:00Z      # Absolute timestamp
```

### Cluster Exploration

```bash
# List indices matching pattern
clibana -u https://logs.internal -i "pods-*" indices

# Show field mappings
clibana -u https://logs.internal -i "pods-*" mappings

# Quiet mode (no headers)
clibana -u https://logs.internal -i "pods-*" indices -q
clibana -u https://logs.internal -i "pods-*" mappings -q
```

### AWS Managed OpenSearch

```bash
# Use aws:// scheme to resolve domain name to endpoint
clibana -u aws://logs-internal -i "logs-*" search "error"

# Override AWS auth with basic auth
clibana -u aws://logs-internal -i "logs-*" -U admin --password-file ~/.clibana-password search "error"
```

## Configuration

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `CLIBANA_URL` | OpenSearch URL (`https://host:port` or `aws://domain-name`) | - |
| `CLIBANA_INDEX` | Index pattern | - |
| `CLIBANA_AUTH` | Authentication type (`aws`, `basic`, or `cookie`) | auto-detected |
| `CLIBANA_COOKIE_FILE` | Path to Netscape cookie file for cookie-based auth | - |
| `CLIBANA_SERVER_TYPE` | Server type for cookie auth/console proxy: `opensearch` or `elasticsearch` | auto-detected |
| `CLIBANA_USER` | Username for basic authentication | - |
| `CLIBANA_FIELDS` | Comma-separated list of fields to output | - |
| `CLIBANA_DEBUG` | Enable debug output | `false` |

### Authentication

**Basic Authentication:**

```bash
echo 'your-password' > ~/.clibana-password
chmod 600 ~/.clibana-password
clibana -u https://logs.internal -i "logs-*" -U admin --password-file ~/.clibana-password search "error"
```

**AWS Authentication:**

AWS credentials are loaded from the default credential chain (environment variables, shared credentials file, IAM role).

**Cookie Authentication:**

Cookie authentication routes requests through the dashboard proxy (OpenSearch Dashboards or Kibana). This is useful when direct API access is not available or requires SSO/SAML authentication.

To export cookies from your browser, use the [Get cookies.txt LOCALLY](https://github.com/kairi003/Get-cookies.txt-LOCALLY) Chrome extension. This creates a standard Netscape cookie file format (compatible with curl, wget, etc.).

```bash
# Use cookie authentication
clibana -u https://dashboards.internal -i "logs-*" -C ~/cookies.txt search "error"

# Specify server type manually if auto-detection fails
clibana -u https://dashboards.internal -i "logs-*" -C ~/cookies.txt -S elasticsearch search "error"
```

Server type auto-detection is used for cookie authentication and console proxy to determine the correct API paths. Use `-S` to override if needed.

## Requirements

### AWS Managed OpenSearch

When using the `aws://` scheme, the following IAM permissions are required:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": ["es:DescribeDomain"],
      "Resource": "arn:aws:es:*:*:domain/*"
    },
    {
      "Effect": "Allow",
      "Action": ["es:ESHttpGet", "es:ESHttpPost"],
      "Resource": "arn:aws:es:*:*:domain/*/_search"
    }
  ]
}
```

## License

[GPL-3.0](LICENSE)
