package togomcp

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

// Validation for caller-supplied generator input (the make_resource tool), so an
// MCP client can't smuggle extra args/flags into the togo subprocess.
var (
	resourceNameRe = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]*$`)
	fieldRe        = regexp.MustCompile(`^[a-z][a-z0-9_]*:[a-z][a-z0-9_]*(:nullable)?$`)
)

// registerTools wires togo's standard MCP toolset (read tools for all roles,
// mutating generators for admin) — the Laravel-Boost-style introspection set.
func (s *Server) registerTools() {
	AddTool(s, "list_resources", "List resources from togo.resources.yaml.", func(_ context.Context, _ NoArgs) (string, error) {
		return readFileOr("togo.resources.yaml", "no resources yet (togo.resources.yaml not found)"), nil
	})

	AddTool(s, "app_info", "Show app info: module, togo version, resource count.", func(_ context.Context, _ NoArgs) (string, error) {
		mod := firstLineWith(readFileOr("go.mod", ""), "module ")
		res := readFileOr("togo.resources.yaml", "")
		return fmt.Sprintf("module: %s\nresources:\n%s", strings.TrimPrefix(mod, "module "), res), nil
	})

	AddTool(s, "get_config", "Read app config (.env with secrets redacted, + togo.yaml).", func(_ context.Context, _ NoArgs) (string, error) {
		return redactEnv(readFileOr(".env", "")) + "\n---\n" + readFileOr("togo.yaml", "(no togo.yaml)"), nil
	})

	AddTool(s, "read_logs", "Tail the application log (LOG_FILE or storage/logs).", func(_ context.Context, in struct {
		Lines int `json:"lines"`
	}) (string, error) {
		if in.Lines <= 0 {
			in.Lines = 100
		}
		return tailLog(in.Lines), nil
	})

	AddTool(s, "db_schema", "List database tables and columns.", func(_ context.Context, _ NoArgs) (string, error) {
		return dbSchema()
	})

	AddTool(s, "db_query", "Run a read-only SELECT against the app database.", func(_ context.Context, in struct {
		Query string `json:"query"`
	}) (string, error) {
		return dbQuery(in.Query)
	})

	AddTool(s, "list_routes", "List the generated REST routes (from the resource manifest).", func(_ context.Context, _ NoArgs) (string, error) {
		return listRoutes(), nil
	})

	if !s.Admin {
		return
	}
	AddTool(s, "make_resource", "Scaffold a full resource (model, REST, GraphQL, page).", func(_ context.Context, in struct {
		Name   string   `json:"name"`
		Fields []string `json:"fields"`
	}) (string, error) {
		if !resourceNameRe.MatchString(in.Name) {
			return "", fmt.Errorf("invalid resource name %q (letters, digits, underscore; start with a letter)", in.Name)
		}
		for _, f := range in.Fields {
			if !fieldRe.MatchString(f) {
				return "", fmt.Errorf("invalid field %q (expected name:type[:nullable])", f)
			}
		}
		return runTogo(append([]string{"make:resource", in.Name}, in.Fields...)...), nil
	})
	AddTool(s, "generate", "Run the codegen pipeline (sqlc -> gqlgen -> atlas -> OpenAPI).", func(_ context.Context, _ NoArgs) (string, error) {
		return runTogo("generate"), nil
	})
	AddTool(s, "migrate", "Apply pending database migrations.", func(_ context.Context, _ NoArgs) (string, error) {
		return runTogo("migrate"), nil
	})
	// Marketplace discovery/install tools (plugins, agents, skills).
	s.registerMarketplaceTools()
}

func runTogo(args ...string) string {
	cmd := exec.Command("togo", args...) //#nosec G204 -- args are togo subcommands from MCP tool input, not a shell
	var buf bytes.Buffer
	cmd.Stdout, cmd.Stderr = &buf, &buf
	cmd.Env = os.Environ()
	_ = cmd.Run()
	return buf.String()
}

func readFileOr(path, fallback string) string {
	b, err := os.ReadFile(path) //#nosec G304,G703 -- reads project config files (go.mod, .env, manifests) in the app dir
	if err != nil {
		return fallback
	}
	return string(b)
}

func firstLineWith(s, prefix string) string {
	for _, ln := range strings.Split(s, "\n") {
		if strings.HasPrefix(strings.TrimSpace(ln), prefix) {
			return strings.TrimSpace(ln)
		}
	}
	return ""
}

// redactEnv hides secret-ish values in a .env dump.
func redactEnv(env string) string {
	var out []string
	for _, ln := range strings.Split(env, "\n") {
		k, _, ok := strings.Cut(ln, "=")
		up := strings.ToUpper(k)
		if ok && (strings.Contains(up, "SECRET") || strings.Contains(up, "PASSWORD") || strings.Contains(up, "KEY") || strings.Contains(up, "TOKEN")) {
			out = append(out, k+"=***redacted***")
		} else {
			out = append(out, ln)
		}
	}
	return strings.Join(out, "\n")
}

func tailLog(n int) string {
	path := os.Getenv("LOG_FILE")
	if path == "" {
		if m, _ := filepath.Glob("storage/logs/*.log"); len(m) > 0 {
			path = m[len(m)-1]
		}
	}
	if path == "" {
		return "(no log file; set LOG_FILE or use storage/logs/*.log)"
	}
	b, err := os.ReadFile(path) //#nosec G304,G703 -- operator-configured log path (LOG_FILE / storage/logs)
	if err != nil {
		return "could not read log: " + err.Error()
	}
	lines := strings.Split(strings.TrimRight(string(b), "\n"), "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return strings.Join(lines, "\n")
}

// openDB opens the app database (sqlite by default) read-only-ish for inspection.
func openDB() (*sql.DB, error) {
	url := os.Getenv("DATABASE_URL")
	if url == "" {
		url = firstLineWith(readFileOr(".env", ""), "DATABASE_URL=")
		url = strings.TrimPrefix(url, "DATABASE_URL=")
	}
	if url == "" {
		url = "file:./togo.db"
	}
	if strings.HasPrefix(url, "postgres") || strings.HasPrefix(url, "pgx") {
		return nil, fmt.Errorf("db inspection currently supports sqlite; got %s", url)
	}
	return sql.Open("sqlite", url)
}

func dbQuery(query string) (string, error) {
	if !strings.HasPrefix(strings.ToUpper(strings.TrimSpace(query)), "SELECT") {
		return "", fmt.Errorf("only SELECT queries are allowed")
	}
	db, err := openDB()
	if err != nil {
		return "", err
	}
	defer db.Close()
	rows, err := db.QueryContext(context.Background(), query)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	cols, _ := rows.Columns()
	var out []map[string]any
	for rows.Next() {
		vals := make([]any, len(cols))
		ptrs := make([]any, len(cols))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return "", err
		}
		m := map[string]any{}
		for i, c := range cols {
			m[c] = vals[i]
		}
		out = append(out, m)
	}
	b, _ := json.MarshalIndent(out, "", "  ")
	return string(b), nil
}

func dbSchema() (string, error) {
	db, err := openDB()
	if err != nil {
		return "", err
	}
	defer db.Close()
	rows, err := db.QueryContext(context.Background(), "SELECT name FROM sqlite_master WHERE type='table' ORDER BY name")
	if err != nil {
		return "", err
	}
	defer rows.Close()
	var sb strings.Builder
	for rows.Next() {
		var t string
		_ = rows.Scan(&t)
		sb.WriteString(t + "\n")
	}
	return sb.String(), nil
}

func listRoutes() string {
	manifest := readFileOr("togo.resources.yaml", "")
	var sb strings.Builder
	for _, ln := range strings.Split(manifest, "\n") {
		ln = strings.TrimSpace(ln)
		if strings.HasPrefix(ln, "table:") {
			table := strings.TrimSpace(strings.TrimPrefix(ln, "table:"))
			fmt.Fprintf(&sb, "GET    /api/%s\nPOST   /api/%s\nGET    /api/%s/{id}\nDELETE /api/%s/{id}\n", table, table, table, table)
		}
	}
	if sb.Len() == 0 {
		return "(no resources; routes appear after make:resource)"
	}
	return sb.String()
}
