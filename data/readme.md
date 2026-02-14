# Welcome to go-mcp

This is a simple MCP (Model Context Protocol) server built with Go.

## Features

- File reading capabilities
- Search functionality  
- List directory contents

## Getting Started

1. Set your GEMINI_API_KEY in .env
2. Run `go run ./cmd/client`
3. Ask questions about the files in the data/ folder

## Project Structure

```
go-mcp/
├── cmd/
│   ├── server/     # MCP server
│   └── client/     # CLI client with Gemini
├── internal/
│   ├── mcp/        # MCP server implementation
│   └── llm/        # Gemini client
└── data/           # Files you can query
```
