# go-gh-scanner

A Go CLI tool that scans GitHub for repositories consuming specific npm dependencies, then searches those repositories' source files for component-level usage. Built for tracking public adoption of a shared design system or library.

## How It Works

The scanner runs in three phases:

1. **Dependency scan** — Searches GitHub for `package.json` files that reference any of the configured npm packages. Optionally scoped to a single org; omit `owner` to search all public repos.

2. **Repository data collection** — For each matching repository, fetches metadata (languages, last push date, default branch) from the GitHub API concurrently.

3. **Component usage search** — Downloads each matching repository's archive and searches its source files for component-level references (Angular selectors, React component names, Vue component names). Component tokens can be provided manually via a JSON file or **auto-discovered** by reading the design system's own source repositories.

Results are written to local JSON files or DynamoDB.

## Authentication

The scanner requires a GitHub personal access token. A **fine-grained PAT** with `Contents: Read` and `Metadata: Read` is sufficient.

### Option 1 — Environment variable (recommended for local use)

```bash
export GITHUB_TOKEN=ghp_yourtoken
```

### Option 2 — AWS Secrets Manager

Set `authTokenKey` in the config to the name of an AWS Secrets Manager secret. The secret must be a JSON object whose key matches the secret name:

```json
{ "my-gh-token-secret": "ghp_yourtoken" }
```

The environment variable takes precedence over Secrets Manager when both are present.

## Configuration

The application is configured via a JSON file passed with the `-c` flag. A complete example is at `config/m2s2-config.json` (file output) or `config/m2s2-dynamo-config.json` (DynamoDB output).

| Field | Required | Description |
|---|---|---|
| `owner` | No | GitHub org slug to restrict the scan to. Omit to search all public repos. |
| `extractDir` | Yes | Temporary directory for downloaded repository archives (e.g. `"temp"`) |
| `perPage` | Yes | Results per page for GitHub API calls (max `100`) |
| `packageFile` | Yes | Package manifest filename to search (e.g. `"package.json"`) |
| `dependencies` | Yes | npm package names to scan for (e.g. `["@m2s2/ng-lib"]`) |
| `languages` | Yes | Source file types to search — list of `{ "name": "...", "extension": "..." }` |
| `reposToIgnore` | No | Repository names to exclude. Supports `*` wildcards (e.g. `"test-*"`) |
| `authTokenKey` | No | AWS Secrets Manager secret name for the GitHub token (omit when using `GITHUB_TOKEN`) |
| `resultsWriterConfig` | Yes | Writer config for all results (see below) |
| `componentDiscovery` | No | Auto-discover component tokens from design system source repos (see below) |

### Writer Config

```json
"resultsWriterConfig": {
  "destinationType": "file",
  "destination": "results",
  "useBatch": false
}
```

| Field | Values | Description |
|---|---|---|
| `destinationType` | `"file"` or `"table"` | Write to a local directory or a DynamoDB table |
| `destination` | path or table name | Root directory for `file`; DynamoDB table name for `table` |
| `useBatch` | `true` / `false` | Use DynamoDB `BatchWriteItem` (only applies to `table`) |

### Component Discovery

When `componentDiscovery` is set, the scanner downloads the design system's own source repositories and extracts component identifiers automatically.

```json
"componentDiscovery": {
  "owner": "m2s2-org",
  "repos": ["m2s2-ng-lib", "m2s2-react-lib", "m2s2-vue-lib"]
}
```

| Field | Description |
|---|---|
| `owner` | GitHub org that owns the design system source repos |
| `repos` | Repository names to scan for component definitions |

Extraction logic per file type:

| Extension | What is extracted |
|---|---|
| `.ts` | Angular `@Component` selector values (e.g. `m2s2-button`) |
| `.tsx` / `.jsx` | Exported PascalCase names (e.g. `M2s2Button`) |
| `.vue` | Component `name:` property values |

When `componentDiscovery` is not set, provide a tokens file via the `-t` flag instead:

```json
["m2s2-button", "m2s2-input", "M2s2Card"]
```

## Results

Both result types are written through the same `resultsWriterConfig`.

### File output

```
results/
  repos/
    m2s2-ng-lib.json
    m2s2-react-lib.json
  components/
    some-org/some-app/
      results_0.json
```

**Repo results** — one JSON file per dependency:

```json
[
  {
    "repo": "some-org/some-app",
    "sk": "DEP#@m2s2/ng-lib",
    "dependency": "@m2s2/ng-lib",
    "version": "2.1.0",
    "url": "https://github.com/some-org/some-app",
    "directory": "packages/web",
    "scm_site": "GitHub",
    "lastModified": "2024-11-01T..."
  }
]
```

**Component results** — one directory per repository:

```json
{
  "repo": "some-org/some-app",
  "sk": "COMP#m2s2-button",
  "component": "m2s2-button",
  "files": [{ "path": "src/app.component.html", "link": "https://github.com/..." }]
}
```

## Running Locally

### Prerequisites

- Go 1.22+
- A GitHub personal access token (fine-grained: `Contents: Read` + `Metadata: Read`)

### Steps

1. **Clone and build:**
   ```bash
   git clone https://github.com/mgmaster24/go-gh-scanner
   cd go-gh-scanner
   go build -o go-gh-scanner .
   ```

2. **Create a config file.** Copy the example and fill in your values:
   ```bash
   cp config/m2s2-config.json config/my-config.json
   ```
   Edit `config/my-config.json`:
   - `componentDiscovery.owner` — the GitHub org that owns your design system source repos
   - `owner` — optional; set to an org slug to limit the scan to one org, or omit to search all public GitHub
   - `resultsWriterConfig.destination` — local directory for output (e.g. `"results"`)

3. **Export your GitHub token:**
   ```bash
   export GITHUB_TOKEN=ghp_yourtoken
   ```

4. **Run:**
   ```bash
   ./go-gh-scanner -c config/my-config.json
   ```
   The `-t` flag is not needed when `componentDiscovery` is configured.

5. **Check the output:**
   ```bash
   ls results/repos/          # one JSON file per dependency
   ls results/components/     # one directory per repo
   ```

### Unit tests

No credentials required:

```bash
go test ./...
```

## Infrastructure

The `infra/` directory contains a CDK app (Go) that provisions everything the scanner needs in AWS:

- **Scanner table** — single DynamoDB table (`repo` PK, `sk` SK) with two GSIs for dependency-first and component-first queries
- **Scanner role** — OIDC-based IAM role the scan workflow assumes; DynamoDB read/write only
- **Infra role** — OIDC-based IAM role the infra deploy workflow assumes; CloudFormation + IAM + DynamoDB

The table uses on-demand billing and a `RETAIN` removal policy so data is never destroyed by a stack update.

### One-time bootstrap

The first deploy must be run locally — the OIDC roles don't exist yet, so there's nothing for the automated workflow to assume.

1. Install prerequisites:
   ```bash
   npm install -g aws-cdk   # CDK CLI
   aws configure            # AWS credentials
   gh auth login            # GitHub CLI
   ```

2. Edit `infra/cdk.json` and fill in the context values:

   | Key | Description |
   |---|---|
   | `tableName` | DynamoDB table name (e.g. `m2s2-scanner`) |
   | `githubRepo` | GitHub repo in `owner/repo` format — scopes the IAM roles |
   | `awsRegion` | AWS region to deploy into (e.g. `us-east-1`) |
   | `scanOwner` | GitHub org to restrict the scan to (leave empty to search all public repos) |
   | `discoveryOwner` | GitHub org that owns the m2s2 source repos |
   | `discoveryRepos` | List of m2s2 source repo names to scan for component definitions |

3. Bootstrap CDK and deploy:
   ```bash
   cd infra
   make bootstrap
   make deploy
   ```

   This provisions the table and IAM roles, writes `config/m2s2-dynamo-config.json`, and sets `AWS_ROLE_ARN`, `AWS_INFRA_ROLE_ARN`, and `AWS_REGION` as GitHub Actions secrets automatically via `gh`.

4. Store your fine-grained GitHub PAT as `SCANNER_GITHUB_TOKEN` in GitHub Actions secrets — this is the only secret the scanner workflow needs that isn't set by `make deploy`.

### Subsequent deploys

After the initial bootstrap, infrastructure changes can be deployed from GitHub without any local tooling: **Actions → Deploy infrastructure → Run workflow**.

### Required GitHub Actions secrets

| Secret | Set by | Description |
|---|---|---|
| `AWS_INFRA_ROLE_ARN` | `make deploy` via `gh` | IAM role with CloudFormation + IAM + DynamoDB permissions for infra deploys |
| `AWS_ROLE_ARN` | `make deploy` via `gh` | IAM role with DynamoDB read/write only for the scanner |
| `AWS_REGION` | `make deploy` via `gh` | AWS region |
| `SCANNER_GITHUB_TOKEN` | You (manually) | Fine-grained PAT with `Contents: Read` and `Metadata: Read` on the m2s2 source org |

## Running on a Schedule

The scanner runs as a GitHub Actions cron job. The workflow at `.github/workflows/scan.yml` triggers every Monday at 06:00 UTC and can be triggered manually from the Actions tab.

## Platform Integration (Admin Dashboard)

For a frontend dashboard, **DynamoDB is recommended** over file output — it lets the frontend query exactly what it needs.

### Table design

One table (`resultsWriterConfig.destination`) holds both dependency and component records:

| Key | Type | Example |
|---|---|---|
| Partition key: `repo` | String | `some-org/some-app` |
| Sort key: `sk` | String | `DEP#@m2s2/ng-lib` or `COMP#m2s2-button` |

Dependency record attributes: `dependency`, `version`, `url`, `directory`, `scm_site`, `lastModified`
Component record attributes: `component`, `files`

**GSIs:**

| Index | PK | SK | Query pattern |
|---|---|---|---|
| `dependency-index` | `dependency` | `repo` | Which repos use `@m2s2/ng-lib`? |
| `component-index` | `component` | `repo` | Which repos use `m2s2-button`? |

The table is fully refreshed on every run — stale dep records for a given repo are deleted before new dep records are written (prefix `DEP#`), and stale component records are deleted before new component records are written (prefix `COMP#`).

## Package Structure

| Package | Purpose |
|---|---|
| `github_api` | GitHub API client — dependency scanning, repository metadata, language lookup, component discovery |
| `config` | Application configuration types and JSON loading |
| `search` | Archive searching — token references in source files, component extraction from library source |
| `tokens` | `TokenReader` interface and implementations (JSON file, Angular component JSON) |
| `models` | Shared data types used across packages |
| `writer` | `ResultsWriter` interface with file and DynamoDB implementations |
| `aws_sdk` | AWS SDK wrappers for Secrets Manager and DynamoDB |
| `cli` | Command-line flag parsing |
| `reader` | Generic JSON file reader |
| `utils` | File I/O, archive extraction, HTTP helpers, string utilities |

## Contributing

```bash
git clone https://github.com/mgmaster24/go-gh-scanner
cd go-gh-scanner
go build
go test ./...
```

Open a pull request to the `main` branch. If you have a question or find a bug, please open an issue.
