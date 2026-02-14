package mcp

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const DataFolder = "data" // base/scoped folder for all file operations.

type ReadFileArgs struct {
	Path string `json:"path" jsonschema:"Path to the file (relative to data/ folder)"`
}

type SearchFilesArgs struct {
	Pattern string `json:"pattern" jsonschema:"Search pattern (regex or text)"`
	Path    string `json:"path" jsonschema:"Directory to search in (relative to data/ folder, default: data/)"`
}

type ListFilesArgs struct {
	Path      string `json:"path" jsonschema:"Directory to list (relative to data/ folder, default: data/)"`
	Recursive bool   `json:"recursive" jsonschema:"List files recursively"`
}

func resolvePath(relativePath string) (string, error) {
	cleanPath := filepath.Clean(relativePath)

	// if the path is empty or just ".", we treat it as the base data folder
	if cleanPath == "" || cleanPath == "." {
		baseDir, _ := filepath.Abs(DataFolder)
		return baseDir, nil
	}
	
	// avoid paths with .. to escape the data folder
	if strings.HasPrefix(cleanPath, "..") {
		return "", fmt.Errorf("path cannot escape data folder")
	}

	baseDir, _ := filepath.Abs(DataFolder)
	resolvedPath, _ := filepath.Abs(filepath.Join(DataFolder, cleanPath))

	if !strings.HasPrefix(resolvedPath, baseDir) {
		return "", fmt.Errorf("path cannot escape data folder")
	}

	return resolvedPath, nil
}

// read a file and return the content as string.
func HandleReadFile(ctx context.Context, req *mcp.CallToolRequest, args ReadFileArgs) (*mcp.CallToolResult, any, error) {
	resolvedPath, err := resolvePath(args.Path)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: " + err.Error()},
			},
			IsError: true,
		}, nil, nil
	}

	// a very basic read file operation.
	content, err := os.ReadFile(resolvedPath)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error reading file: " + err.Error()},
			},
			IsError: true,
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(content)},
		},
	}, nil, nil
}

// search for a pattern in files using grep and return the matching lines.
func HandleSearchFiles(ctx context.Context, req *mcp.CallToolRequest, args SearchFilesArgs) (*mcp.CallToolResult, any, error) {
	resolvedPath, err := resolvePath(args.Path)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: " + err.Error()},
			},
			IsError: true,
		}, nil, nil
	}

	// a recusive grep search to find matching patterns in files.
	cmd := exec.Command("grep", "-r", "-n", "-I", args.Pattern, resolvedPath)
	output, err := cmd.CombinedOutput()

	if err != nil {
		if len(output) == 0 {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "No matches found"},
				},
			}, nil, nil
		}
		// if error is not about no matches found, return the error message.
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: string(output)},
			},
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(output)},
		},
	}, nil, nil
}

// list files in a directory and return the list as string.
func HandleListFiles(ctx context.Context, req *mcp.CallToolRequest, args ListFilesArgs) (*mcp.CallToolResult, any, error) {
	resolvedPath, err := resolvePath(args.Path)
	if err != nil {
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: "Error: " + err.Error()},
			},
			IsError: true,
		}, nil, nil
	}

	var files []string

	// recursive case
	if args.Recursive {
		err = filepath.Walk(resolvedPath, func(p string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			relPath, _ := filepath.Rel(resolvedPath, p)
			files = append(files, relPath)
			return nil
		})
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Error: " + err.Error()},
				},
				IsError: true,
			}, nil, nil
		}
	} else {
		// non-recursive case
		entries, err := os.ReadDir(resolvedPath)
		if err != nil {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "Error: " + err.Error()},
				},
				IsError: true,
			}, nil, nil
		}
		for _, e := range entries {
			files = append(files, e.Name())
		}
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("Files in %s:\n%s", args.Path, strings.Join(files, "\n"))},
		},
	}, nil, nil
}
