// Command mcp is the togo Model Context Protocol server. It exposes togo's
// generators and Laravel-Boost-style introspection as MCP tools, scoped by
// TOGO_MCP_ROLE (admin|user). Runs over stdio by default, or Streamable HTTP when
// MCP_HTTP_ADDR is set (e.g. ":8089"). Thin wrapper over the togomcp SDK.
package main

import (
	"os"

	"github.com/togo-framework/mcp/togomcp"
)

func main() {
	s := togomcp.Default(os.Getenv("TOGO_MCP_ROLE"))
	var err error
	if addr := os.Getenv("MCP_HTTP_ADDR"); addr != "" {
		err = s.RunHTTP(addr)
	} else {
		err = s.Run()
	}
	if err != nil {
		_, _ = os.Stderr.WriteString("mcp: " + err.Error() + "\n")
		os.Exit(1)
	}
}
