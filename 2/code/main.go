package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/googleapis/mcp-toolbox-sdk-go/core"
	"google.golang.org/genai"
)

// ConvertToGenaiTool translates a ToolboxTool into the genai.FunctionDeclaration format.
func ConvertToGenaiTool(toolboxTool *core.ToolboxTool) *genai.Tool {

	inputschema, err := toolboxTool.InputSchema()
	if err != nil {
		return &genai.Tool{}
	}

	var schema *genai.Schema
	_ = json.Unmarshal(inputschema, &schema)
	// First, create the function declaration.
	funcDeclaration := &genai.FunctionDeclaration{
		Name:        toolboxTool.Name(),
		Description: toolboxTool.Description(),
		Parameters:  schema,
	}

	// Then, wrap the function declaration in a genai.Tool struct.
	return &genai.Tool{
		FunctionDeclarations: []*genai.FunctionDeclaration{funcDeclaration},
	}
}

// printResponse extracts and prints the relevant parts of the model's response.
func printResponse(resp *genai.GenerateContentResponse) {
	for _, cand := range resp.Candidates {
		if cand.Content != nil {
			for _, part := range cand.Content.Parts {
				fmt.Println(part.Text)
			}
		}
	}
}

func main() {
	ctx := context.Background()
	apiKey := os.Getenv("GEMINI_API_KEY")
	toolboxURL := "http://localhost:5000"

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		APIKey: apiKey,
	})
	if err != nil {
		log.Fatalf("Failed to create Google GenAI client: %v", err)
	}

	// Initialize the MCP Toolbox client.
	toolboxClient, err := core.NewToolboxClient(toolboxURL)
	if err != nil {
		log.Fatalf("Failed to create Toolbox client: %v", err)
	}

	// Load the tools using the MCP Toolbox SDK.
	tools, err := toolboxClient.LoadToolset("hotel", ctx)
	if err != nil {
		log.Fatalf("Failed to load tools: %v", err)
	}

	genAITools := make([]*genai.Tool, len(tools))
	toolsMap := make(map[string]*core.ToolboxTool, len(tools))

	for i, tool := range tools {
		genAITools[i] = ConvertToGenaiTool(tool)
		toolsMap[tool.Name()] = tool
	}

	modelName := "gemini-2.0-flash"

	query := `You are a hotel expert: 
	You have been given a csv files containing hotel information
	
	Find from the list hotels that is in USA. Only return the hotel name, currency and country
	`

	// Create the initial content prompt for the model.
	contents := []*genai.Content{
		genai.NewContentFromText(query, genai.RoleUser),
	}
	config := &genai.GenerateContentConfig{
		Tools: genAITools,
		ToolConfig: &genai.ToolConfig{
			FunctionCallingConfig: &genai.FunctionCallingConfig{
				Mode: genai.FunctionCallingConfigModeAny,
			},
		},
	}
	genContentResp, _ := client.Models.GenerateContent(ctx, modelName, contents, config)

	printResponse(genContentResp)

	functionCalls := genContentResp.FunctionCalls()
	if len(functionCalls) == 0 {
		log.Println("No function call returned by the AI. The model likely answered directly.")
		return
	}

	// Process the first function call (the example assumes one for simplicity).
	fc := functionCalls[0]
	log.Printf("--- Gemini requested function call: %s ---\n", fc.Name)
	log.Printf("--- Arguments: %+v ---\n", fc.Args)

	var toolResultString string

	if fc.Name == "search-hotels-by-name" {
		tool := toolsMap["search-hotels-by-name"]
		toolResult, err := tool.Invoke(ctx, fc.Args)
		toolResultString = fmt.Sprintf("%v", toolResult)
		if err != nil {
			log.Fatalf("Failed to execute tool '%s': %v", fc.Name, err)
		}

	} else {
		log.Println("LLM did not request our tool")
	}
	resultContents := []*genai.Content{
		genai.NewContentFromText("You are given the following hotel data in CSV format:"+toolResultString+
			"Task: Answer questions by filtering this dataset based on the criteria provided."+
			"Find all hotels that are located in USA."+
			"Only return the hotel name, rating and country", genai.RoleUser),
	}

	finalResponse, err := client.Models.GenerateContent(ctx, modelName, resultContents, &genai.GenerateContentConfig{})
	if err != nil {
		log.Fatalf("Error calling GenerateContent (with function result): %v", err)
	}
	log.Println("=== Final Response from Model (after processing function result) ===")
	printResponse(finalResponse)

}
