# CloudNecromancer

Reconstruct point-in-time AWS infrastructure snapshots by replaying CloudTrail events.

Given any historical timestamp, CloudNecromancer resurrects every resource that existed at that moment — EC2 instances, IAM roles, S3 buckets, Lambda functions, security groups, VPCs, RDS databases, and more — from create/modify/delete event chains stored in CloudTrail.

```
 ░░░░░░░░░░░░░░░░░░░░░░░░░░
 ░  ☠  CloudNecromancer  ☠  ░
 ░░░░░░░░░░░░░░░░░░░░░░░░░░
 Raising the dead since 2026
```

## Use Cases

- **Incident response** — "What was running at 3am before the breach?"
- **Compliance audits** — Point-in-time inventory for any past date
- **Post-incident timelines** — Full infrastructure state reconstruction
- **Drift analysis** — Compare two timestamps to see what changed

## Install

```bash
# Homebrew (macOS / Linux)
brew tap pfrederiksen/tap
brew install cloudnecromancer

# From source (requires CGO + DuckDB headers)
go install github.com/pfrederiksen/cloudnecromancer@latest
```

Pre-built binaries for macOS (Intel/Apple Silicon) and Linux (amd64) are available on the [Releases](https://github.com/pfrederiksen/cloudnecromancer/releases) page.

## Quick Start

```bash
# 1. Fetch CloudTrail events into a local database
cloudnecromancer fetch \
  --account-id 123456789012 \
  --regions us-east-1,us-west-2 \
  --start 2026-01-01T00:00:00Z \
  --end 2026-03-01T00:00:00Z

# 2. Resurrect infrastructure at a specific point in time
cloudnecromancer resurrect --at 2026-02-15T03:00:00Z

# 3. Compare two points in time
cloudnecromancer diff \
  --from 2026-01-01T00:00:00Z \
  --to 2026-02-15T03:00:00Z

# 4. Export as Terraform HCL
cloudnecromancer resurrect --at 2026-02-15T03:00:00Z --format terraform --output snapshot.tf
```

## Data Sources

### CloudTrail API (default)

Queries the `LookupEvents` API directly. Simple to set up, but limited to the last 90 days.

```bash
cloudnecromancer fetch \
  --source cloudtrail \
  --account-id 123456789012 \
  --regions us-east-1,us-west-2 \
  --start 2026-01-01T00:00:00Z \
  --end 2026-03-01T00:00:00Z \
  [--profile my-aws-profile]
```

Requires read-only CloudTrail permissions. See [AWS_PERMISSIONS.md](AWS_PERMISSIONS.md) for the minimal IAM policy.

### Splunk

Queries CloudTrail events from a Splunk index. No time limit — go as far back as your retention allows.

```bash
cloudnecromancer fetch \
  --source splunk \
  --splunk-url https://splunk.corp:8089 \
  --splunk-token $SPLUNK_TOKEN \
  --account-id 123456789012 \
  --regions us-east-1,us-west-2 \
  --start 2024-01-01T00:00:00Z \
  --end 2026-03-01T00:00:00Z
```

| Flag | Default | Description |
|------|---------|-------------|
| `--splunk-url` | *(or `SPLUNK_URL` env)* | Splunk REST API base URL (e.g. `https://splunk.corp:8089`) |
| `--splunk-token` | *(or `SPLUNK_TOKEN` env)* | Splunk bearer token |
| `--splunk-index` | `aws_cloudtrail` | Splunk index containing CloudTrail events |
| `--splunk-sourcetype` | `aws:cloudtrail` | Splunk sourcetype |
| `--splunk-query` | *(auto-generated)* | Override the generated SPL query entirely |
| `--splunk-insecure` | `false` | Skip TLS certificate verification |

The auto-generated SPL query searches the configured index/sourcetype and filters by account ID and regions. Use `--splunk-query` to provide your own SPL if your CloudTrail data uses a non-standard schema.

## Commands

| Command | Description |
|---------|-------------|
| `fetch` | Pull CloudTrail events from CloudTrail API or Splunk into local DuckDB |
| `resurrect` | Reconstruct infrastructure state at a point in time |
| `diff` | Compare infrastructure between two timestamps |
| `export` | Re-export an existing snapshot in a different format |
| `info` | Show database statistics |

### `resurrect`

```bash
cloudnecromancer resurrect \
  --at 2026-02-15T03:00:00Z \
  [--services ec2,iam,s3] \
  [--region us-east-1] \
  [--format json|terraform|cloudformation|cdk|pulumi|ocsf|csv] \
  [--output ./snapshot.json] \
  [--include-dead] \
  [--ritual]
```

### `diff`

```bash
cloudnecromancer diff \
  --from 2026-01-01T00:00:00Z \
  --to 2026-02-15T03:00:00Z \
  [--format table|json]
```

### `export`

```bash
cloudnecromancer export \
  --input ./snapshot.json \
  --format hcl \
  --output ./snapshot.tf
```

## Output Formats

| Format | Flag | Description |
|--------|------|-------------|
| JSON | `--format json` | Full snapshot with nested resource attributes |
| Terraform | `--format terraform` | HCL with `import` + `resource` blocks (aliases: `hcl`, `tf`) |
| CloudFormation | `--format cloudformation` | CloudFormation JSON template (alias: `cfn`) |
| CDK | `--format cdk` | AWS CDK TypeScript stack using L1 constructs |
| Pulumi | `--format pulumi` | Pulumi TypeScript program using `@pulumi/aws` |
| OCSF | `--format ocsf` | OCSF Inventory Info events (class_uid 5001), NDJSON |
| CSV | `--format csv` | Splunk lookup table format |

<details>
<summary>Output format examples</summary>

**JSON** (`--format json`):

```json
{
  "timestamp": "2026-02-15T03:00:00Z",
  "account_id": "123456789012",
  "resources": {
    "ec2:instance": [
      {
        "resource_id": "i-0abc123def456789",
        "state": "running",
        "attributes": {
          "instance_type": "t3.medium",
          "image_id": "ami-0abcdef1234567890"
        }
      }
    ]
  }
}
```

**Terraform HCL** (`--format terraform`):

```hcl
import {
  to = aws_instance.i_0abc123def456789
  id = "i-0abc123def456789"
}

resource "aws_instance" "i_0abc123def456789" {
  instance_type = "t3.medium"
  ami           = "ami-0abcdef1234567890"
  subnet_id     = "subnet-0123456789abcdef0"
}
```

**CDK** (`--format cdk`):

```typescript
new ec2.CfnInstance(this, "i-0abc123def456789", {
  instanceType: "t3.medium",
  imageId: "ami-0abcdef1234567890",
  subnetId: "subnet-0123456789abcdef0",
});
```

**Pulumi** (`--format pulumi`):

```typescript
const i_0abc123def456789 = new aws.ec2.Instance("i-0abc123def456789", {
    instanceType: "t3.medium",
    ami: "ami-0abcdef1234567890",
    subnetId: "subnet-0123456789abcdef0",
});
```

**CSV** (`--format csv`) — for use as a Splunk lookup table:

```spl
| inputlookup cloudnecromancer_lookup.csv
| join resource_id [search index=cloudtrail earliest=-30d]
```

</details>

## Global Flags

| Flag | Default | Description |
|------|---------|-------------|
| `--db` | `./necromancer.db` | Path to DuckDB database file |
| `--profile` | *(default chain)* | AWS profile to use |
| `--quiet` | `false` | Suppress banner and non-essential output |
| `--verbose` | `false` | Enable verbose logging |

## Supported Services (23)

| Service | Resources |
|---------|-----------|
| EC2 | instances, VPCs, subnets, security groups, IGWs |
| IAM | roles, users, policies |
| S3 | buckets, policies, versioning, public access |
| Lambda | functions |
| RDS | instances, clusters |
| ELB | load balancers, target groups, listeners |
| ECS | clusters, services, task definitions |
| EKS | clusters, nodegroups |
| KMS | keys, aliases, rotation |
| Secrets Manager | secrets, rotation, restore |
| CloudWatch Logs | log groups, log streams, retention |
| DynamoDB | tables, global tables |
| SNS | topics, subscriptions |
| SQS | queues, queue attributes |
| API Gateway | REST APIs, HTTP APIs, stages |
| Route 53 | hosted zones, record sets |
| ECR | repositories, lifecycle policies, scanning |
| ElastiCache | clusters, replication groups |
| WAF v2 | web ACLs, rule groups, IP sets |
| GuardDuty | detectors, filters |
| CloudFront | distributions, origin access controls |
| EBS | volumes, snapshots |
| SSM | documents, parameters, maintenance windows |

All services support create, update, and delete event tracking (133 CloudTrail events total).

## How It Works

1. **Fetch** — Pull CloudTrail events from the CloudTrail API or Splunk and store them in an embedded DuckDB database. Multi-region fetches run concurrently.

2. **Parse** — Each event is routed through a service-specific parser that extracts a `ResourceDelta`: the action (create/update/delete), resource ID, and relevant attributes.

3. **Replay** — To reconstruct state at time T, the engine queries all events before T (ordered chronologically) and applies each delta to an in-memory resource map.

4. **Export** — The final snapshot is serialized in the requested format.

## Development

```bash
make build    # Build binary to ./bin/cloudnecromancer
make test     # Run all tests
make lint     # Run golangci-lint
```

### Adding a new service parser

1. Create `internal/parser/services/myservice.go`
2. Implement the `Parser` interface
3. Register in `init()` with `parser.Register(&MyServiceParser{})`
4. Add test fixtures to `testdata/`
5. Add table-driven tests in `internal/parser/services/myservice_test.go`

### Adding a new exporter

1. Create `internal/export/myformat.go`
2. Implement the `Exporter` interface (`Export(snapshot, writer) error`)
3. Register it in `GetExporter()` in `internal/export/exporter.go`
4. Add tests in `internal/export/export_test.go`

## License

MIT
