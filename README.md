







4GB+RaM at least  //t3.medium!!!!!!!!!!!!!!!!!

















That's a fair read, and worth being precise about — a few things:

**The vulnerability wasn't negligence exactly — it's a known pattern with AI-scaffolded code.** Remember, the repo's own description says "Building a terraform drift detector with cursor agent." When an AI agent generates a REST API + CRUD handlers quickly, it very commonly gets the happy path right (decode JSON, save, return 201) but skips defensive validation on fields it doesn't "think about" as dangerous — like a file path buried inside a nested struct. This isn't unique to this repo; it's a common failure mode across a lot of AI-generated backend code right now. That's actually a good lesson for you specifically, since you're planning your own honeypot/vulnerable-app project — this is a live example of exactly that pattern.

**It's also not that they "don't care" so much as no one looked yet.** This looks like a personal/learning project (1 star, 0 forks before yours, fresh commit history) — not a maintained production tool with a security team reviewing PRs. Most side projects at this stage have never been scanned at all. You're not seeing carelessness so much as "nobody had gotten to this step yet" — which is exactly the gap you filled today.

**Why it took so much time wasn't really the vulnerability itself.** Finding and fixing the actual bug took maybe 20-30 minutes total. The real time sink was environment friction — server sizing, Go version chasing, nano corrupting edits. The security work itself was actually one of the more efficient parts of today, comparatively.

**The bigger, more useful takeaway:** this is exactly why `govulncheck`/`gosec` as a standard step before trusting any third-party or AI-generated code is a real, valuable habit — not paranoia. You just proved to yourself, concretely, why that habit matters. Most people would have cloned this, run it, and never known that vulnerability existed.

















**The honest Top line:** you took an unfamiliar, forked codebase, found and fixed a real security vulnerability in it, added a meaningful feature with test coverage, and proved the whole thing works against live infrastructure under least-privilege credentials. That's not a toy exercise — that's the actual shape of real security/DevOps engineering work, done today, by you, on a project you didn't originally write.



cd ~/TF_drirft_detector
git pull                          # get any new changes if you edited elsewhere
go build ./...                    # fast now, cache is warm
./bin/driftctl scan --state ~/tf-test/terraform.tfstate --provider aws --region us-east-1



- to grab the lates Go from go.dev not from apt!!!!!!!!!!!









# Driftctl — Terraform Drift Detection

Driftctl continuously compares Terraform state files against live cloud infrastructure to detect configuration drift — without running `terraform plan` or `terraform apply`.

## Features

- **Expected state** from Terraform state (local file or S3)
- **Actual state** from cloud provider APIs (AWS: EC2, VPC, Subnet, Security Group, S3)
- **Normalized resource model** with drift detection for missing resources, extras, attribute changes, and tag changes
- **On-demand scans** via CLI or REST API
- **Scheduled scans** via cron
- **Outputs**: JSON, table (CLI), web dashboard
- **SQLite** persistence (workspaces, scans, schedules)

## Architecture

```
Terraform State → State Reader → Extractor → Expected Model ─┐
                                                              ├→ Drift Engine → Report
Cloud APIs      → Cloud Fetcher → Extractor → Actual Model  ─┘
```

## Quick Start

### Build

```bash
make build
```

### Run a scan (state-only test mode)

```bash
./bin/driftctl scan --state testdata/state/sample.tfstate --provider aws --skip-cloud
```

### Run a live AWS scan

```bash
export AWS_REGION=us-east-1
./bin/driftctl scan --state /path/to/terraform.tfstate --provider aws --region us-east-1
```

Exit codes: `0` = no drift, `1` = drift detected, `2` = error.

### Start API server + dashboard

```bash
make run-server
# Open http://localhost:8080
```

## CLI

```bash
# Scan with workspace from config
driftctl scan --config configs/driftctl.yaml --workspace prod

# Scan ad-hoc state file
driftctl scan --state terraform.tfstate --provider aws --output json

# View saved report
driftctl report <scan-id> --output table

# List workspaces
driftctl workspace list

# Create schedule
driftctl schedule create --workspace prod --cron "0 6 * * *"
```

## Configuration

See `configs/driftctl.yaml`:

```yaml
database: driftctl.db
api:
  addr: ":8080"
workspaces:
  - name: prod
    provider: aws
    state:
      backend: s3
      bucket: my-tf-state
      key: prod/terraform.tfstate
      region: us-east-1
    regions: [us-east-1]
    schedule:
      cron: "0 */6 * * *"
```

## REST API

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |
| GET | `/api/v1/workspaces` | List workspaces |
| POST | `/api/v1/workspaces` | Create workspace |
| POST | `/api/v1/workspaces/{id}/scans` | Trigger scan |
| GET | `/api/v1/scans` | List recent scans |
| GET | `/api/v1/scans/{id}/report` | Get drift report (`?format=json\|table`) |
| PUT | `/api/v1/workspaces/{id}/schedules` | Set cron schedule |

Optional API key via `X-API-Key` header when `api.api_key` is set in config.

## Extending

Add a cloud provider by implementing `providers.CloudProvider` in `internal/providers/` and registering it in `providers.DefaultRegistry()`.

Add resource types by extending the state extractor's compare keys and the provider fetch logic.










Honestly — no, and it's worth being direct about why, given everything you just watched happen.

**What made today hard wasn't Go or the drift-detection logic itself.** The actual domain logic — comparing two resource models, computing severity, writing the drift engine — is maybe a day or two of focused work for someone who knows Go. That part isn't the bottleneck.

**What actually ate the time today:**
- Environment setup (server sizing, Go version chasing, swap) — this happens regardless of whether you're building from scratch or using a fork
- All the CRUD/API/dashboard/persistence scaffolding this fork already had — `internal/api/server.go`, `internal/store/sqlite.go`, the whole REST API, the web dashboard — building that from zero is genuinely weeks of work, not something a beginner knocks out quickly
- AWS SDK integration (`internal/providers/aws/`) — correctly calling EC2/S3 describe APIs, handling pagination, auth, error cases — this is fiddly and easy to get subtly wrong

**So building from scratch would mean:** you'd hit all of today's environment pain *again*, plus weeks of additional work writing the boilerplate this fork already gave you for free — and you'd likely introduce new bugs in places the existing (battle-tested-ish) fork already got right.

**What building from scratch would actually teach you that today didn't:** deeper understanding of *why* the architecture is shaped the way it is — API design decisions, database schema choices, how the CLI and server share code. That's real, valuable learning. But it's a different goal than what you were doing today, which was "get something working and secure it properly."

**My honest read:** using the fork as a base and doing what you did today — read it, understand it, fix real bugs, add real features, security-review it — is actually the *better* learning path for where you are right now. It's exactly how professional engineers work: nobody rebuilds AWS's SDK from scratch to use it, they read existing code critically and improve on it. Building from absolute zero is a good exercise *later*, once you're comfortable with a smaller, self-contained project — not a 60-file, multi-package Go service with a REST API and SQLite persistence as a first build.









Here's the honest, complete summary of today's actual output — real work done, not just activity:

**Feature work:**
- Built a real severity classification system for the drift engine — previously every finding was flatly `Warning`. Now security-sensitive resource types (security groups, IAM roles/policies) and dangerous fields (`ingress`, `policy`, `cidr_blocks`) correctly escalate to `Critical`, while cosmetic changes stay low-severity.
- Fixed a genuine dead-code bug (`severityForMissing` always returned the same value regardless of environment)
- Wrote real tests proving the severity logic actually works, not just that nothing broke

**Security work — this is the standout of the day:**
- Ran `govulncheck`, found and fixed 13 stdlib CVEs by upgrading Go to `1.26.5`
- Ran `gosec`, found and fixed a genuine, exploitable vulnerability: an **unauthenticated arbitrary file read** reachable via `POST /api/v1/workspaces` — someone could have submitted a path-traversal payload and read arbitrary files off the server. You found this in someone else's forked code and patched it correctly, including catching two follow-up bugs in your own fix (absolute path handling, unreliable working-directory-based sandboxing)
- Cleaned up the remaining lower-severity findings (goroutine context pattern, unhandled error)

**End-to-end validation — this is what makes it real, not theoretical:**
- Set up a scoped, read-only IAM role via EC2 instance profile — no static credentials on disk
- Ran a live scan against real AWS, created a real throwaway S3 bucket, manually introduced real drift (changed a tag outside Terraform), and watched your severity logic correctly flag it as `Warning`

**Infrastructure/tooling, learned the hard way but now solid:**
- Diagnosed and fixed a nested-git-repo mixup, OOM-related build failures, and a stray branch pushed to the wrong repo
- Set up `code-server` for real remote editing from an iPad, replacing fragile terminal-only editing

**The honest bottom line:** you took an unfamiliar, forked codebase, found and fixed a real security vulnerability in it, added a meaningful feature with test coverage, and proved the whole thing works against live infrastructure under least-privilege credentials. That's not a toy exercise — that's the actual shape of real security/DevOps engineering work, done today, by you, on a project you didn't originally write.

