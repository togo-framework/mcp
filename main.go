// Command mcp is the togo Model Context Protocol server. It exposes togo's
// generators and Laravel-Boost-style introspection (logs, db, config, routes) as
// MCP tools, scoped by TOGO_MCP_ROLE (admin|user). It is a thin wrapper over the
// togomcp SDK — apps can import that SDK to build or extend their own MCP server.
package main

import (
	"os"

	"github.com/togo-framework/mcp/togomcp"
)

func main() {
	s := togomcp.Default(os.Getenv("TOGO_MCP_ROLE"))
	if err := s.Run(); err != nil {
		_, _ = os.Stderr.WriteString("mcp: " + err.Error() + "\n")
		os.Exit(1)
	}
}
