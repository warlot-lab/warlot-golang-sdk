package devcli

import (
	"encoding/json"
	"fmt"
	"os"
)

// PrintJSON prints a value as pretty-printed JSON.
func PrintJSON(v any) {
	b, _ := json.MarshalIndent(v, "", "  ")
	fmt.Println(string(b))
}

// PrintGlobalUsage renders the top-level usage text.
func PrintGlobalUsage(bin string) {
	// Environment defaults echoed inline for transparency.
	fmt.Println(bin + ` - official CLI for the Warlot SQL API

USAGE:
  ` + bin + ` <command> [flags]

GLOBAL FLAGS (env defaults shown in []):
  -base          	API base URL [` + getenvDefault(EnvBaseURL, "https://api.warlot.stevenhert.xyz") + `]
  -apikey        	API key [` + getenvDefault(EnvAPIKey, "") + `]
  -holder        	Holder ID [` + getenvDefault(EnvHolderID, "") + `]
  -pname         	Project name [` + getenvDefault(EnvProjectName, "") + `]
  -timeout       	Request timeout seconds [` + getenvDefault(EnvTimeoutSec, "90") + `]
  -retries       	Retries on 429/5xx [` + getenvDefault(EnvRetries, "5") + `]
  -backoff-init  	Initial backoff ms [` + getenvDefault(EnvBackoffInit, "1000") + `]
  -backoff-max   	Max backoff ms [` + getenvDefault(EnvBackoffMax, "8000") + `]
  -v             	Verbose logs

COMMANDS:
  resolve             	                          Resolve project by holder + name
  init        		-owner 0x...                  Initialize new project
  issue-key   		-project <id> -user 0x...     Issue a project API key

  sql         		-project <id> -q "SQL ..." [-params '[...]' -idempotency key -stream]
  tables list   	-project <id>
  tables browse 	-project <id> -table products [-limit 10 -offset 0]
  tables schema 	-project <id> -table products
  tables count  	-project <id>
  status      		-project <id>
  commit      		-project <id>
  ready       		                              Check cluster readiness probe
  indexer     		                              Query indexer projections (storage, balances, history)

EXAMPLES:
  ` + bin + ` status -project <id>
  ` + bin + ` sql -project <id> -q 'SELECT * FROM products ORDER BY id DESC LIMIT 5'
  ` + bin + ` tables schema -project <id> -table products
  ` + bin + ` ready
`)
}

// ExitWithUsageError formats a usage error to stderr and exits with status 2.
func ExitWithUsageError(err error, hint string) {
	fmt.Fprintf(os.Stderr, "error: %v\n", err)
	if hint != "" {
		fmt.Fprintf(os.Stderr, "hint: %s\n", hint)
	}
	os.Exit(2)
}
