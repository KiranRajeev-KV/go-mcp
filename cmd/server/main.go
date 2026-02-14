package main

import (
	"context"
	"fmt"
	"os"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	mcpserver "github.com/user/go-mcp/internal/mcp"
)

func main() {
	server := mcpserver.NewServer()

	if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
		fmt.Fprintln(os.Stderr, "Server error:", err)
		os.Exit(1)
	}
}
