package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/modelcontextprotocol/go-sdk/mcp"
	"google.golang.org/genai"
)

// ToolConverter converts MCP tools to Gemini API tools and extracts results from tool calls.
type ToolConverter struct{}

// NewToolConverter creates a new instance of ToolConverter.
func NewToolConverter() *ToolConverter {
	return &ToolConverter{}
}

// the tools in MCP and gemini have diffrent structures (yes this is a pain and i dont know why they are not standardized)
// for expample, read_file in MCP has name, description and input schema
// but when converting to Gemini tool, we need to create a FunctionDeclaration with name, description and parameters
// so we need to convert the MCP tool structure to Gemini tool structure
func (tc *ToolConverter) MCPToGeminiToolsSlice(tools []*mcp.Tool) []*genai.Tool {
	var geminiTools []*genai.Tool

	for _, tool := range tools {
		inputSchema := tc.convertSchema(tool.InputSchema)

		decl := &genai.FunctionDeclaration{
			Name:        tool.Name,
			Description: tool.Description,
			Parameters:  inputSchema,
		}
		geminiTools = append(geminiTools, &genai.Tool{FunctionDeclarations: []*genai.FunctionDeclaration{decl}})
	}

	return geminiTools
}

// converts the input schema from MCP to Gemini format. 
// This is a recursive function to handle nested schemas.
func (tc *ToolConverter) convertSchema(schema any) *genai.Schema {
	if schema == nil {
		return &genai.Schema{
			Type:       genai.TypeObject,
			Properties: map[string]*genai.Schema{},
		}
	}

	switch s := schema.(type) {
	case map[string]any:
		return tc.mapToSchema(s)
	default:
		return &genai.Schema{
			Type:       genai.TypeObject,
			Properties: map[string]*genai.Schema{},
		}
	}
}


// convert a map[string]any to genai.Schema recursively
// the schema can be nested, so we do it recursively.
func (tc *ToolConverter) mapToSchema(m map[string]any) *genai.Schema {
	schema := &genai.Schema{
		Type:       genai.TypeObject,
		Properties: map[string]*genai.Schema{},
	}

	if t, ok := m["type"].(string); ok {
		schema.Type = genai.Type(t)
	}

	if props, ok := m["properties"].(map[string]any); ok {
		for key, val := range props {
			if propMap, ok := val.(map[string]any); ok {
				schema.Properties[key] = tc.mapToSchema(propMap)
			}
		}
	}

	if required, ok := m["required"].([]any); ok {
		for _, r := range required {
			if rStr, ok := r.(string); ok {
				schema.Required = append(schema.Required, rStr)
			}
		}
	}

	return schema
}

// takes the result of a tool call and get the text from it.
// assumption here is that the tool result is text content and returns a string.
func (tc *ToolConverter) ExtractToolResult(result *mcp.CallToolResult) string {
	if result == nil {
		return ""
	}

	var text strings.Builder
	for _, content := range result.Content {
		if tc, ok := content.(*mcp.TextContent); ok {
			text.WriteString(tc.Text)
		}
	}
	return text.String()
}

type GeminiClient struct {
	client *genai.Client
	model  string
}

// makes a new Gemini client :)
func NewGeminiClient(apiKey string) (*GeminiClient, error) {
	client, err := genai.NewClient(context.Background(), &genai.ClientConfig{
		APIKey:  apiKey,
		Backend: genai.BackendGeminiAPI,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return &GeminiClient{
		client: client,
		model:  "gemini-2.5-flash-lite",
	}, nil
}

type FunctionCall struct {
	Name      string
	Arguments map[string]any
}

// generate content with system prompt, user prompt and tools - single turn, no history
func (g *GeminiClient) GenerateContent(ctx context.Context, systemPrompt, userPrompt string, tools []*genai.Tool) (*genai.GenerateContentResponse, error) {
	contents := []*genai.Content{
		{
			Role:  "user",
			Parts: []*genai.Part{{Text: userPrompt}},
		},
	}

	config := &genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{
			Role:  "model",
			Parts: []*genai.Part{{Text: systemPrompt}},
		},
		Tools: tools,
	}

	return g.client.Models.GenerateContent(ctx, g.model, contents, config)
}

// generate content with system prompt, user prompt, tools and history - multi turn with history
func (g *GeminiClient) GenerateContentWithHistory(ctx context.Context, history []*genai.Content, tools []*genai.Tool) (*genai.GenerateContentResponse, error) {
	config := &genai.GenerateContentConfig{
		Tools: tools,
	}

	return g.client.Models.GenerateContent(ctx, g.model, history, config)
}

// generate content with system prompt, user prompt, tools and history - multi turn with history and system prompt
func (g *GeminiClient) GenerateContentWithSystemPrompt(ctx context.Context, systemPrompt string, history []*genai.Content, tools []*genai.Tool) (*genai.GenerateContentResponse, error) {
	config := &genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{
			Role:  "model",
			Parts: []*genai.Part{{Text: systemPrompt}},
		},
		Tools: tools,
	}

	return g.client.Models.GenerateContent(ctx, g.model, history, config)
}

// extract function(tool) calls from response and then excute the tool and get the result,
// then add the function call and response to history and generate again with the new history
// until there is no more function call
func ExtractFunctionCalls(resp *genai.GenerateContentResponse) []FunctionCall {
	var calls []FunctionCall

	for _, candidate := range resp.Candidates {
		if candidate.Content == nil {
			continue
		}
		for _, part := range candidate.Content.Parts {
			if part.FunctionCall != nil {
				args := make(map[string]any)
				if part.FunctionCall.Args != nil {
					jsonBytes, _ := json.Marshal(part.FunctionCall.Args)
					json.Unmarshal(jsonBytes, &args)
				}
				calls = append(calls, FunctionCall{
					Name:      part.FunctionCall.Name,
					Arguments: args,
				})
			}
		}
	}

	return calls
}

// simplest function here :cry - just extract the text from the response and return it as a string
func ExtractText(resp *genai.GenerateContentResponse) string {
	var text strings.Builder

	for _, candidate := range resp.Candidates {
		if candidate.Content == nil {
			continue
		}
		for _, part := range candidate.Content.Parts {
			if part.Text != "" {
				text.WriteString(part.Text)
			}
		}
	}

	return text.String()
}
