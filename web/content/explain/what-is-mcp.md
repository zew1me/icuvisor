---
title: "What is MCP?"
description: "A short explanation of Model Context Protocol for icuvisor users."
---

Model Context Protocol, or MCP, is a standard way for an AI client to talk to local or remote tools. In icuvisor's case, the AI client starts the local icuvisor server over stdio or connects to an already running Streamable HTTP server, asks which tools are available, and then calls those tools when you ask questions about your intervals.icu data.

That means you can use one local icuvisor install with MCP-compatible clients such as Claude Desktop, Claude Code, ChatGPT, Cursor, Continue, Zed, Pi, and others as their MCP support evolves. The official MCP project is at <https://modelcontextprotocol.io>.

icuvisor does not replace intervals.icu and does not train a model. It is the protocol layer that lets your chosen AI assistant request structured data from intervals.icu through a local process you control.
