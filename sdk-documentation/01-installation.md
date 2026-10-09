# Installation

Official Go SDK for the Warlot SQL Database API. This page covers prerequisites, module installation, environment configuration, a minimal verification run, and core type references relevant to setup.

---

## Prerequisites

| Requirement  | Details                                                                                 |
| ------------ | --------------------------------------------------------------------------------------- |
| Go toolchain | Go **1.18+** (generic functions are used)                                               |
| Platforms    | Linux, macOS, Windows (including WSL)                                                   |
| Network      | HTTPS egress to the configured API base URL (default `https://warlot-api.onrender.com`) |

---

## Module installation

### Add to a project (Go modules)

```bash
# inside a Go module
go get github.com/warlot-lab/warlot-golang-sdk/warlot-go/warlot@latest
```

### Pin to a specific version

```bash
go get github.com/warlot-lab/warlot-golang-sdk/warlot-go/warlot@v0.1.0
```

### Update to latest compatible

```bash
go get -u github.com/warlot-lab/warlot-golang-sdk/warlot-go/warlot
```

---

## Import path and minimal construction

```go
import "github.com/warlot-lab/warlot-golang-sdk/warlot-go/warlot"

func initClient() *warlot.Client {
    return warlot.New(
        warlot.WithHolderID("0x..."),
        warlot.WithProjectName("project_name"),
        // API key may be attached later via cl.APIKey = "…"
    )
}
```

---

## Optional environment configuration

The SDK and the CLI (`warlotctl`) honor these environment variables when present.

| Variable                 | Purpose                        | Example                           |
| ------------------------ | ------------------------------ | --------------------------------- |
| `WARLOT_BASE_URL`        | Override API base URL          | `https://warlot-api.onrender.com` |
| `WARLOT_API_KEY`         | Default API key header         | `a2f5…37e0`                       |
| `WARLOT_HOLDER`          | Default holder identifier      | `0x2e4a…7ba3`                     |
| `WARLOT_PNAME`           | Default project name           | `my_project`                      |
| `WARLOT_TIMEOUT`         | HTTP request timeout (seconds) | `90`                              |
| `WARLOT_RETRIES`         | Max retries on 429/5xx         | `5`                               |
| `WARLOT_BACKOFF_INIT_MS` | Initial backoff (milliseconds) | `1000`                            |
| `WARLOT_BACKOFF_MAX_MS`  | Max backoff (milliseconds)     | `8000`                            |

---

## Minimal verification (smoke test)

### Compile check

```bash
go env -w GOFLAGS=-mod=mod
go list -m github.com/warlot-lab/warlot-golang-sdk/warlot-go/warlot
```

### Unit tests (offline)

```bash
# from repository root or module where tests are included
go test ./warlot -v
```

### Live E2E (optional)

```bash
export WARLOT_E2E=1
export WARLOT_HOLDER=0x...     # holder address
export WARLOT_OWNER=0x...      # owner/user address
export WARLOT_PNAME=my_project # project name

# optional base override:
# export WARLOT_BASE_URL=https://warlot-api.onrender.com

go test ./e2e -v
```

---

## Types referenced during installation

These definitions are provided to clarify constructor and option usage during setup. A complete type catalog is available in `12-types.md`.

```go
// New constructs a Client with safe defaults.
// Options may override base URL, headers, timeouts, backoff, and logging hooks.
func New(opts ...Option) *Client

// Option customizes a Client at construction time.
type Option func(*Client)

// Common options for initial setup.
func WithBaseURL(u string) Option
func WithAPIKey(key string) Option
func WithHolderID(holder string) Option
func WithProjectName(name string) Option
func WithHTTPClient(h *http.Client) Option
func WithUserAgent(ua string) Option
func WithRetries(max int) Option
func WithBackoff(initial, max time.Duration) Option
func WithLogger(l Logger) Option

// Client includes shared configuration and HTTP plumbing.
type Client struct {
    BaseURL       string
    APIKey        string
    HolderID      string
    ProjectName   string
    HTTPClient    *http.Client
    UserAgent     string
    MaxRetries    int
    InitialBackoff time.Duration
    MaxBackoff     time.Duration
    Logger        Logger
    BeforeHooks   []func(*http.Request)
    AfterHooks    []func(*http.Response, []byte, error)
}
```

---

## Test definitions (installation context)

* **Unit tests** reside alongside the SDK (`warlot/*_test.go`) and rely on stubbed servers and fixtures under `warlot/testdata/`. Execution target: `go test ./warlot -v`.
* **Live E2E** resides in `e2e/` and exercises the hosted API; requires environment variables (`WARLOT_E2E`, `WARLOT_HOLDER`, `WARLOT_OWNER`, `WARLOT_PNAME`). Execution target: `go test ./e2e -v`.

Detailed guidance is documented in `13-testing.md`.

---

## Troubleshooting (installation)

| Symptom                               | Cause                                                       | Resolution                                                                         |
| ------------------------------------- | ----------------------------------------------------------- | ---------------------------------------------------------------------------------- |
| `no test files` under `./warlot`      | Tests placed under `testdata/` instead of package directory | Move `*_test.go` into `warlot/`; keep `testdata/` for fixtures only                |
| `stat …/example: directory not found` | Attempt to run a non-existent example path                  | Use CLI or unit tests instead; reference `11-cli.md`                               |
| 404 with JSON body via curl           | Upstream proxy/header mismatch                              | Use SDK or pass `Accept: application/json`; retry via SDK to observe parsed fields |
| TLS or network timeout                | Corporate proxy/firewall or WSL DNS                         | Configure proxy env (`HTTPS_PROXY`), confirm DNS, increase `WithBackoff`/timeouts  |

For additional issues, refer to `14-troubleshooting.md` and `15-security.md`.
