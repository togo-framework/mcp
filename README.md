<!-- togo-brand -->
<p align="center">
  <img src=".github/assets/togo-mark.svg" width="96" alt="togo" />
</p>
<h1 align="center">togo-mcp</h1>
<p align="center"><sub>part of the <a href="https://github.com/togo-framework">togo-framework</a> — the full-stack Go + React framework</sub></p>

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


---

## 💎 Premium sponsors

togo is proudly sponsored by **ID8 Media** and **One Studio**.

<p align="center">
  <a href="https://id8media.com"><img src=".github/assets/id8media.svg" height="44" alt="ID8 Media" /></a>
  &nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;
  <a href="https://one-studio.co"><img src=".github/assets/one-studio.jpeg" height="44" alt="One Studio" /></a>
</p>
