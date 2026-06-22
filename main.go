// Command mcp is the Model Context Protocol server for the togo framework.
// It exposes togo's generators and project introspection as MCP tools so AI
// agents (Claude Code, Cursor, …) can drive a togo project end-to-end.
//
// It runs over stdio and shells out to the `togo` CLI in the current project
// directory. Wire it into an agent with `togo mcp:install --agent claude-code`.
package main

import (
	"bytes"
	"context"
	"os"
	"os/exec"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// noArgs is the input type for tools that take no arguments.
type noArgs struct{}

// output is the common structured result: the captured CLI output.
type output struct {
	Output string `json:"output" jsonschema:"the command output"`
}

type makeResourceArgs struct {
	Name   string   `json:"name" jsonschema:"resource name in PascalCase, e.g. Post"`
	Fields []string `json:"fields" jsonschema:"fields as name:type, e.g. title:string body:text:nullable"`
}

func main() {
	server := mcp.NewServer(&mcp.Implementation{
		Name:    "togo",
		Version: "0.1.0",
	}, nil)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "make_resource",
		Description: "Scaffold a full resource (model, sqlc, Atlas, GraphQL, REST, seeder, Next.js page) in the current togo project.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, in makeResourceArgs) (*mcp.CallToolResult, output, error) {
		args := append([]string{"make:resource", in.Name}, in.Fields...)
		return run(args...)
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "generate",
		Description: "Run the togo codegen pipeline (sqlc → gqlgen → atlas diff → OpenAPI export).",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ noArgs) (*mcp.CallToolResult, output, error) {
		return run("generate")
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "list_resources",
		Description: "List the resources defined in the project's togo.resources.yaml manifest.",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ noArgs) (*mcp.CallToolResult, output, error) {
		data, err := os.ReadFile("togo.resources.yaml")
		if err != nil {
			return text("no resources yet (togo.resources.yaml not found)")
		}
		return text(string(data))
	})

	mcp.AddTool(server, &mcp.Tool{
		Name:        "migrate",
		Description: "Apply pending database migrations (Atlas).",
	}, func(ctx context.Context, _ *mcp.CallToolRequest, _ noArgs) (*mcp.CallToolResult, output, error) {
		return run("migrate")
	})

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		os.Stderr.WriteString("mcp: " + err.Error() + "\n")
		os.Exit(1)
	}
}

// run executes `togo <args...>` in the current directory and returns its output.
func run(args ...string) (*mcp.CallToolResult, output, error) {
	cmd := exec.Command("togo", args...)
	var buf bytes.Buffer
	cmd.Stdout, cmd.Stderr = &buf, &buf
	cmd.Env = os.Environ()
	_ = cmd.Run() // include output even on non-zero exit
	return text(buf.String())
}

func text(s string) (*mcp.CallToolResult, output, error) {
	return &mcp.CallToolResult{
		Content: []mcp.Content{&mcp.TextContent{Text: s}},
	}, output{Output: s}, nil
}
