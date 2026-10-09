# CLI (`warlotdev`)


## A) Linux/macOS (bash/zsh)

### 1) Install the CLI from the published submodule tag

```bash
go install github.com/warlot-lab/warlot-golang-sdk/warlot-go/cmd/warlotdev@v1.0.1
```

### 2) Ensure the binary is on PATH (current shell + future sessions)

```bash
export PATH="$(go env GOBIN):$(go env GOPATH)/bin:$PATH"
printf '\nexport PATH="$(go env GOBIN):$(go env GOPATH)/bin:$PATH"\n' >> ~/.bashrc 2>/dev/null || true
printf '\nexport PATH="$(go env GOBIN):$(go env GOPATH)/bin:$PATH"\n' >> ~/.zshrc  2>/dev/null || true
hash -r 2>/dev/null || true
warlotdev -h
```

### 3) Baseline configuration (environment defaults)

```bash
export WARLOT_BASE_URL="https://warlot-api.onrender.com"
export WARLOT_HOLDER="REPLACE_WITH_HOLDER_ID"
export WARLOT_PNAME="REPLACE_WITH_PROJECT_NAME"
export WARLOT_TIMEOUT="90"
export WARLOT_RETRIES="6"
export WARLOT_BACKOFF_INIT_MS="500"
export WARLOT_BACKOFF_MAX_MS="8000"
```

### 4) Project bootstrap: resolve → init (if needed) → issue API key

```bash
set -euo pipefail

# Resolve by (holder, project name)
RESOLVE_JSON="$(warlotdev -base "$WARLOT_BASE_URL" -holder "$WARLOT_HOLDER" -pname "$WARLOT_PNAME" resolve)"
echo "$RESOLVE_JSON" | jq -C .

# Extract project id; support legacy fields if present
PROJECT_ID="$(echo "$RESOLVE_JSON" | jq -r '.project_id // .ProjectID // empty')"

# Initialize if not present
if [ -z "${PROJECT_ID}" ] ; then
  OWNER_ADDR="REPLACE_WITH_OWNER_ADDRESS"
  INIT_JSON="$(warlotdev -base "$WARLOT_BASE_URL" -holder "$WARLOT_HOLDER" -pname "$WARLOT_PNAME" init -owner "$OWNER_ADDR")"
  echo "$INIT_JSON" | jq -C .
  PROJECT_ID="$(echo "$INIT_JSON" | jq -r '.ProjectID')"
fi

# Issue API key bound to the project
USER_ADDR="REPLACE_WITH_USER_ADDRESS"
ISSUE_JSON="$(warlotdev -base "$WARLOT_BASE_URL" -holder "$WARLOT_HOLDER" -pname "$WARLOT_PNAME" issue-key -project "$PROJECT_ID" -user "$USER_ADDR")"
echo "$ISSUE_JSON" | jq -C .
export WARLOT_API_KEY="$(echo "$ISSUE_JSON" | jq -r '.apiKey')"

echo "PROJECT_ID=$PROJECT_ID"
echo "WARLOT_API_KEY set"
```

### 5) Basic operations (schema + data + status + commit)

```bash
# Create a table
warlotdev -apikey "$WARLOT_API_KEY" -holder "$WARLOT_HOLDER" -pname "$WARLOT_PNAME" \
  sql -project "$PROJECT_ID" \
  -q 'CREATE TABLE IF NOT EXISTS products (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, price REAL)'

# Insert with idempotency (safe on retries)
warlotdev -apikey "$WARLOT_API_KEY" -holder "$WARLOT_HOLDER" -pname "$WARLOT_PNAME" \
  sql -project "$PROJECT_ID" \
  -q 'INSERT INTO products (name, price) VALUES (?, ?)' \
  -params '["Laptop", 999.99]' \
  -idempotency 'cli-insert-001'

# Query rows
warlotdev -apikey "$WARLOT_API_KEY" -holder "$WARLOT_HOLDER" -pname "$WARLOT_PNAME" \
  sql -project "$PROJECT_ID" \
  -q 'SELECT id, name, price FROM products ORDER BY id'

# List tables
warlotdev -apikey "$WARLOT_API_KEY" -holder "$WARLOT_HOLDER" -pname "$WARLOT_PNAME" \
  tables list -project "$PROJECT_ID"

# Browse with pagination
warlotdev -apikey "$WARLOT_API_KEY" -holder "$WARLOT_HOLDER" -pname "$WARLOT_PNAME" \
  tables browse -project "$PROJECT_ID" -table products -limit 10 -offset 0

# Project status
warlotdev -apikey "$WARLOT_API_KEY" -holder "$WARLOT_HOLDER" -pname "$WARLOT_PNAME" \
  status -project "$PROJECT_ID"

# Commit changes to chain-backed storage
warlotdev -apikey "$WARLOT_API_KEY" -holder "$WARLOT_HOLDER" -pname "$WARLOT_PNAME" \
  commit -project "$PROJECT_ID"
```

### 6) Optional maintenance

```bash
# Verbose diagnostics (redacts API key)
warlotdev -v -base "$WARLOT_BASE_URL" -holder "$WARLOT_HOLDER" -pname "$WARLOT_PNAME" resolve

# Reinstall after a fresh release
go clean -modcache
go install github.com/warlot-lab/warlot-golang-sdk/warlot-go/cmd/warlotdev@v1.0.1
```

---

## B) Windows (PowerShell)

### 1) Install the CLI from the published submodule tag

```powershell
go install github.com/warlot-lab/warlot-golang-sdk/warlot-go/cmd/warlotdev@v1.0.1
```

### 2) Ensure the binary is on PATH (persistent for User; current session updated)

```powershell
$bin = (go env GOBIN)
if (!$bin) { $bin = (Join-Path (go env GOPATH) 'bin') }

# Persist PATH (User)
$current = [Environment]::GetEnvironmentVariable('PATH','User')
if ($current -notlike "*$bin*") {
  [Environment]::SetEnvironmentVariable('PATH', "$bin;$current", 'User')
}
# Update current session
$env:PATH = "$bin;$env:PATH"

warlotdev -h
```

### 3) Baseline configuration (environment defaults)

```powershell
$env:WARLOT_BASE_URL        = "https://warlot-api.onrender.com"
$env:WARLOT_HOLDER          = "REPLACE_WITH_HOLDER_ID"
$env:WARLOT_PNAME           = "REPLACE_WITH_PROJECT_NAME"
$env:WARLOT_TIMEOUT         = "90"
$env:WARLOT_RETRIES         = "6"
$env:WARLOT_BACKOFF_INIT_MS = "500"
$env:WARLOT_BACKOFF_MAX_MS  = "8000"
```

### 4) Project bootstrap: resolve → init (if needed) → issue API key

```powershell
# Resolve by (holder, project name)
$res = warlotdev -base $env:WARLOT_BASE_URL -holder $env:WARLOT_HOLDER -pname $env:WARLOT_PNAME resolve
$res | Out-String

# Extract project id; support legacy keys
$projectId = ($res | ConvertFrom-Json).project_id
if (-not $projectId) { $projectId = ($res | ConvertFrom-Json).ProjectID }

# Initialize if not present
if (-not $projectId) {
  $owner = "REPLACE_WITH_OWNER_ADDRESS"
  $init = warlotdev -base $env:WARLOT_BASE_URL -holder $env:WARLOT_HOLDER -pname $env:WARLOT_PNAME init -owner $owner
  $projectId = ($init | ConvertFrom-Json).ProjectID
}

# Issue API key bound to the project
$user = "REPLACE_WITH_USER_ADDRESS"
$issue = warlotdev -base $env:WARLOT_BASE_URL -holder $env:WARLOT_HOLDER -pname $env:WARLOT_PNAME issue-key -project $projectId -user $user
$env:WARLOT_API_KEY = ($issue | ConvertFrom-Json).apiKey

"PROJECT_ID=$projectId"
"Set WARLOT_API_KEY"
```

### 5) Basic operations (schema + data + status + commit)

```powershell
# Create a table
warlotdev -apikey $env:WARLOT_API_KEY -holder $env:WARLOT_HOLDER -pname $env:WARLOT_PNAME `
  sql -project $projectId `
  -q 'CREATE TABLE IF NOT EXISTS products (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT, price REAL)'

# Insert with idempotency
warlotdev -apikey $env:WARLOT_API_KEY -holder $env:WARLOT_HOLDER -pname $env:WARLOT_PNAME `
  sql -project $projectId `
  -q 'INSERT INTO products (name, price) VALUES (?, ?)' `
  -params '["Laptop", 999.99]' `
  -idempotency 'cli-insert-001'

# Query rows
warlotdev -apikey $env:WARLOT_API_KEY -holder $env:WARLOT_HOLDER -pname $env:WARLOT_PNAME `
  sql -project $projectId `
  -q 'SELECT id, name, price FROM products ORDER BY id'

# List tables
warlotdev -apikey $env:WARLOT_API_KEY -holder $env:WARLOT_HOLDER -pname $env:WARLOT_PNAME `
  tables list -project $projectId

# Browse with pagination
warlotdev -apikey $env:WARLOT_API_KEY -holder $env:WARLOT_HOLDER -pname $env:WARLOT_PNAME `
  tables browse -project $projectId -table products -limit 10 -offset 0

# Project status
warlotdev -apikey $env:WARLOT_API_KEY -holder $env:WARLOT_HOLDER -pname $env:WARLOT_PNAME `
  status -project $projectId

# Commit changes
warlotdev -apikey $env:WARLOT_API_KEY -holder $env:WARLOT_HOLDER -pname $env:WARLOT_PNAME `
  commit -project $projectId
```

### 6) Optional maintenance

```powershell
# Verbose diagnostics (redacts API key)
warlotdev -v -base $env:WARLOT_BASE_URL -holder $env:WARLOT_HOLDER -pname $env:WARLOT_PNAME resolve

# Reinstall after a fresh release
go clean -modcache
go install github.com/warlot-lab/warlot-golang-sdk/warlot-go/cmd/warlotdev@v1.0.1
```

---

**Placeholders to replace before use**

* `REPLACE_WITH_HOLDER_ID` – chain holder identifier
* `REPLACE_WITH_PROJECT_NAME` – project name label
* `REPLACE_WITH_OWNER_ADDRESS` – owner address for initialization
* `REPLACE_WITH_USER_ADDRESS` – user address for API key issuance

These batches provide end-to-end setup and operation for **warlotdev** with consistent configuration across environments.
