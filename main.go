// Command togo-mcp is the MCP server for the togo framework. It exposes togo's
// generators and project introspection as MCP tools so AI agents (Claude Code,
// Cursor, …) can drive a togo project end-to-end.
//
// Planned tools:
//   - make_resource      scaffold a resource (model + sqlc + Atlas + GraphQL + REST)
//   - generate           run the codegen pipeline
//   - list_resources     read togo.resources.yaml
//   - migrate / seed     database lifecycle
//   - install_plugin     add a plugin from a GitHub repo
//
// Served over streamable HTTP at /mcp/admin and /mcp/user with role-based access,
// wired into an agent via `togo mcp:install --agent claude-code`.
package main

import "fmt"

func main() {
	fmt.Println("togo-mcp: MCP server (implementation lands in the MCP phase)")
}
