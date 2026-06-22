// Package togomcp is the togo MCP SDK. Apps embed it to build their own MCP
// server or extend togo's: Default(role) returns a server preloaded with togo's
// tools (generators + Laravel-Boost-style introspection), and AddTool lets you
// register custom tools before Run.
//
//	s := togomcp.Default(os.Getenv("TOGO_MCP_ROLE"))
//	togomcp.AddTool(s, "my_tool", "does X", func(ctx, in MyArgs) (string, error) { ... })
//	s.Run()
package togomcp

import (
	"context"
	"net/http"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	_ "modernc.org/sqlite"
)

// Output is the structured result type shared by togo tools.
type Output struct {
	Output string `json:"output"`
}

// Server wraps an MCP server with togo conveniences.
type Server struct {
	MCP   *mcp.Server
	Admin bool
}

// New builds a bare server (no tools registered).
func New(name, version string) *Server {
	return &Server{MCP: mcp.NewServer(&mcp.Implementation{Name: name, Version: version}, nil)}
}

// Default returns a togo server preloaded with the standard toolset, scoped by
// role ("user" = read-only; anything else = admin with mutating generators).
func Default(role string) *Server {
	s := New("togo", "0.4.0")
	s.Admin = role != "user"
	s.registerTools()
	return s
}

// Run serves over stdio (default transport for local agents).
func (s *Server) Run() error {
	return s.MCP.Run(context.Background(), &mcp.StdioTransport{})
}

// RunHTTP serves over Streamable HTTP at addr (e.g. ":8089") for remote agents.
func (s *Server) RunHTTP(addr string) error {
	handler := mcp.NewStreamableHTTPHandler(func(*http.Request) *mcp.Server { return s.MCP }, nil)
	srv := &http.Server{Addr: addr, Handler: handler, ReadHeaderTimeout: 10 * time.Second}
	return srv.ListenAndServe()
}

// AddTool registers a typed tool (apps use this to extend the server).
func AddTool[In any](s *Server, name, description string, fn func(context.Context, In) (string, error)) {
	mcp.AddTool(s.MCP, &mcp.Tool{Name: name, Description: description},
		func(ctx context.Context, _ *mcp.CallToolRequest, in In) (*mcp.CallToolResult, Output, error) {
			out, err := fn(ctx, in)
			if err != nil {
				return text(err.Error())
			}
			return text(out)
		})
}

func text(s string) (*mcp.CallToolResult, Output, error) {
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: s}}}, Output{Output: s}, nil
}

// NoArgs is the input for tools that take no arguments.
type NoArgs struct{}
