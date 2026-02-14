package mcp

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

func NewServer() *mcp.Server {
	s := mcp.NewServer(&mcp.Implementation{
		Name:    "go-mcp-tools",
		Version: "1.0.0",
	}, nil)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "read_file",
		Description: "Read the contents of a file from the local filesystem",
	}, HandleReadFile)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "search_files",
		Description: "Search for text patterns in files using grep",
	}, HandleSearchFiles)

	mcp.AddTool(s, &mcp.Tool{
		Name:        "list_files",
		Description: "List files in a directory",
	}, HandleListFiles)

	return s
}
