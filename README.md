# go-mcp

An MCP (Model Context Protocol) server with Gemini LLM integration, built with Go.

## Overview

This project demonstrates a working MCP server that provides file operations tools (read, search, list) and integrates with Google's Gemini for natural language queries. The MCP server runs as a subprocess and communicates via stdio transport.

## Features

- **MCP Server** - Provides 3 file operations tools:
  - `read_file` - Read file contents from the data/ folder
  - `search_files` - Search for patterns in files using grep
  - `list_files` - List files in a directory

- **Gemini Integration** - Natural language interface using Google Gemini API
- **Scoped Access** - File operations are restricted to the `data/` folder for security
- **Interactive CLI** - Query your files using natural language

## Prerequisites

- Go 1.24+
- Gemini API key

## Setup

1. **Get a Gemini API key**:
   - Go to [Google AI Studio](https://aistudio.google.com/app/apikey)
   - Create a new API key
   - Copy it to `.env`:

2. **Create `.env` file**:
   ```bash
   cp .env.example .env
   ```
   
   Or manually create `.env`:
   ```
   GEMINI_API_KEY=your_api_key_here
   ```

## Usage

```bash
go run ./cmd/client
```

Then type your queries. Examples:

```
>> list the files in the data folder
>> read team.json
>> what products are in stock?
>> search for "ERROR" in the logs
>> find functions in calculator.go
```

Type `quit` or `exit` to stop.

## Project Structure

```
go-mcp/
├── cmd/
│   ├── server/         # MCP server entry point
│   └── client/         # CLI client with Gemini
├── internal/
│   ├── mcp/            # MCP server implementation
│   │   ├── server.go    # Server setup
│   │   └── tools.go    # Tool handlers
│   └── llm/            # Gemini client
│       └── client.go    # LLM integration
├── data/               # Files you can query (scoped access)
├── .env                # API keys (gitignored)
├── .env.example        # Template for .env
├── go.mod
└── README.md
```

## Data Folder

The `data/` folder contains sample files you can query:

| File | Description |
|------|-------------|
| `readme.md` | Project documentation |
| `team.json` | Team members with roles & skills |
| `products.csv` | Products with prices & stock |
| `calculator.go` | Go calculator code |
| `process_data.py` | Python data processing |
| `todo.txt` | Project TODO list |
| `server.log` | Sample server logs |

## How It Works

1. **Client starts** → Spawns MCP server as subprocess
2. **List tools** → Client fetches available tools from server
3. **Convert schemas** → MCP tool definitions → Gemini function declarations
4. **Interactive loop**:
   - You type query → Gemini decides if tools needed
   - If yes → Client calls MCP tool → Returns result to Gemini → Gemini responds
   - If no → Gemini responds directly

## Tech Stack

- [Go MCP SDK](https://github.com/modelcontextprotocol/go-sdk) - Official MCP protocol implementation
- [Google GenAI SDK](https://pkg.go.dev/google.golang.org/genai) - Gemini API client

## License

MIT
