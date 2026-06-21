# togo-mcp

The MCP server for [togo](https://github.com/togo-framework/togo). It exposes
togo's generators and project introspection as Model Context Protocol tools so
AI agents — Claude Code, Cursor, Windsurf, Cline — can scaffold resources, run
codegen, manage migrations, and install plugins inside a togo project.

Wire it into your agent with:

```bash
togo mcp:install --agent claude-code --role admin
```

It serves over streamable HTTP at `/mcp/admin` and `/mcp/user` with role-based
access. Every `togo new` project also ships a `.mcp.json` pre-wired to this
server.

## License

MIT
