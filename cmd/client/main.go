package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/joho/godotenv"
	"github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/user/go-mcp/internal/llm"
	"google.golang.org/genai"
)

const systemPrompt = `You are a helpful AI assistant. You have access to tools that can read files, search for patterns in files, and list files in the data/ folder only.

Available tools:
- read_file: Read the contents of a file. Input is the file path relative to data/ (e.g., "team.json" or "subfolder/file.txt")
- search_files: Search for text patterns in files. Input is the pattern to search for.
- list_files: List files in a directory. Input is the directory path relative to data/ (default: empty = data folder)

IMPORTANT: All file paths should be relative to data/ folder. For example:
- To read data/team.json, use path: "team.json"
- To search in data folder, leave path empty
- NEVER use absolute paths like /home/...

When you need to use a tool, make the function call and I will execute it and return the results. Then you can provide your answer based on the results.

If you don't need any tools, just answer the question directly.`

func main() {
	godotenv.Load()

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "Error: GEMINI_API_KEY environment variable is not set")
		fmt.Fprintln(os.Stderr, "Please set it with: export GEMINI_API_KEY=your_key")
		os.Exit(1)
	}

	ctx := context.Background()

	// Start MCP server as a subprocess
	fmt.Println("Starting MCP server...")
	serverCmd := exec.Command("go", "run", "./cmd/server")
	serverCmd.Stderr = os.Stderr

	transport := &mcp.CommandTransport{Command: serverCmd}

	client := mcp.NewClient(&mcp.Implementation{
		Name:    "mcp-client",
		Version: "1.0.0",
	}, nil)

	session, err := client.Connect(ctx, transport, nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error connecting to MCP server:", err)
		os.Exit(1)
	}
	defer session.Close()

	fmt.Println("Fetching available tools...")
	// get the list of tools from the MCP server
	toolResult, err := session.ListTools(ctx, &mcp.ListToolsParams{})
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error listing tools:", err)
		os.Exit(1)
	}


	// Convert MCP tools to Gemini tools format
	toolConverter := llm.NewToolConverter()
	geminiTools := toolConverter.MCPToGeminiToolsSlice(toolResult.Tools)

	fmt.Printf("Found %d tools: ", len(toolResult.Tools))
	for _, t := range toolResult.Tools {
		fmt.Printf("%s ", t.Name)
	}
	fmt.Println()

	geminiClient, err := llm.NewGeminiClient(apiKey)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating Gemini client:", err)
		os.Exit(1)
	}

	// a very simple in-memory history for the session
	var history []*genai.Content

	fmt.Println("\n=== MCP CLI with Gemini ===")
	fmt.Println("Type your queries below. Type 'quit' or 'exit' to exit.")

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print(">> ")
		query, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading input:", err)
			continue
		}

		query = strings.TrimSpace(query)
		if query == "" {
			continue
		}

		if query == "quit" || query == "exit" {
			fmt.Println("Goodbye!")
			break
		}

		// Add user message to history
		history = append(history, &genai.Content{
			Role:  "user",
			Parts: []*genai.Part{{Text: query}},
		})

		// Generate content
		resp, err := geminiClient.GenerateContentWithSystemPrompt(ctx, systemPrompt, history, geminiTools)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error generating content:", err)
			continue
		}

		// Check for function calls
		functionCalls := llm.ExtractFunctionCalls(resp)

		for len(functionCalls) > 0 {
			fmt.Printf("[Calling tool: %s]\n", functionCalls[0].Name)

			// Execute tool call via MCP
			toolResp, err := session.CallTool(ctx, &mcp.CallToolParams{
				Name:      functionCalls[0].Name,
				Arguments: functionCalls[0].Arguments,
			})
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error calling tool: %v\n", err)
				break
			}

			toolResultText := toolConverter.ExtractToolResult(toolResp)
			fmt.Printf("[Tool result: %s]\n", truncate(toolResultText, 100)) // truncate long results for display

			// Add function call and response to history
			history = append(history, &genai.Content{
				Role: "model",
				Parts: []*genai.Part{
					{
						FunctionCall: &genai.FunctionCall{
							Name: functionCalls[0].Name,
							Args: functionCalls[0].Arguments,
						},
					},
				},
			})

			history = append(history, &genai.Content{
				Role: "user",
				Parts: []*genai.Part{
					{
						FunctionResponse: &genai.FunctionResponse{
							Name: functionCalls[0].Name,
							Response: map[string]any{
								"result": toolResultText,
							},
						},
					},
				},
			})

			// Generate again with tool result
			resp, err = geminiClient.GenerateContentWithSystemPrompt(ctx, systemPrompt, history, geminiTools)
			if err != nil {
				fmt.Fprintln(os.Stderr, "Error generating content:", err)
				break
			}

			functionCalls = llm.ExtractFunctionCalls(resp)
		}

		// Extract and display final response
		responseText := llm.ExtractText(resp)
		if responseText != "" {
			fmt.Println(responseText)

			// Add assistant response to history
			history = append(history, &genai.Content{
				Role:  "model",
				Parts: []*genai.Part{{Text: responseText}},
			})
		}
		fmt.Println()
	}
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
