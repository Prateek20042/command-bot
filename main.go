package main

import (
	"bufio"
	"command-bot/logger"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/go-resty/resty/v2"
)

type AnalysisResult struct {
	Instructions     []string               `json:"instructions"`
	Actions          []string               `json:"actions"`
	ResolvedEntities map[string]interface{} `json:"resolved_entities"`
}

type OllamaResponse struct {
	Response string `json:"response"`
}

func main() {
	logger.InitLogger("bot.log")
	defer logger.InfoLog.Println("Session ended")

	fmt.Println("\nGo-chatbot")

	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("command: ")
		userInput, _ := reader.ReadString('\n')
		userInput = strings.TrimSpace(userInput)

		if userInput == "exit" {
			break
		}

		if err := processCommand(userInput); err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}
}

func processCommand(input string) error {
	analysis, err := getCommandAnalysis(input)
	if err != nil {
		return err
	}
	displayResults(input, analysis)
	return nil
}

func getCommandAnalysis(input string) (*AnalysisResult, error) {
	prompt := fmt.Sprintf(`Analyze this work command: "%s"

Extract and infer:
1. Single most important instruction
2. 2 most relevant actions (verb+object)
3. Resolved entities with department inference

For people, return EXACTLY:
{
  "name": "full name",
  "department": "specific department (infer from context)",
  "position": "current position",
  "traits": ["key characteristics"]
}

Department inference rules:
- "selling target" → Sales
- "customer" → Sales
- "marketing" → Marketing
- "code/technical" → Engineering
- "finance" → Finance
- Default: "Department unspecified"

Return ONLY this JSON format:
{
  "instructions": ["instruction"],
  "actions": ["action1", "action2"],
  "resolved_entities": {
    "person_name": {
      "name": "name",
      "department": "specific department",
      "position": "position",
      "traits": ["trait1", "trait2"]
    }
  }
}`, input)

	resp, err := resty.New().R().
		SetHeader("Content-Type", "application/json").
		SetBody(map[string]interface{}{
			"model":  "llama3",
			"prompt": prompt,
			"format": "json",
			"stream": false,
		}).
		Post("http://localhost:11434/api/generate")

	if err != nil {
		return nil, fmt.Errorf("API call failed: %w", err)
	}

	var ollamaResp OllamaResponse
	if err := json.Unmarshal(resp.Body(), &ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	jsonStr := extractJSON(ollamaResp.Response)
	if jsonStr == "" {
		return nil, fmt.Errorf("no valid JSON found in response")
	}

	var analysis AnalysisResult
	if err := json.Unmarshal([]byte(jsonStr), &analysis); err != nil {
		return nil, fmt.Errorf("failed to parse analysis: %w\nResponse was: %s", err, ollamaResp.Response)
	}

	analysis = cleanAnalysis(analysis, input)
	return &analysis, nil
}

func cleanAnalysis(analysis AnalysisResult, input string) AnalysisResult {

	if len(analysis.Instructions) > 0 {
		analysis.Instructions = []string{analysis.Instructions[0]}
	}

	var cleanedActions []string
	actionSet := make(map[string]bool)
	for _, action := range analysis.Actions {
		if !isGenericAction(action) && !actionSet[action] {
			cleanedActions = append(cleanedActions, action)
			actionSet[action] = true
		}
	}
	if len(cleanedActions) > 2 {
		cleanedActions = cleanedActions[:2]
	}
	analysis.Actions = cleanedActions

	for _, details := range analysis.ResolvedEntities {
		if entityMap, ok := details.(map[string]interface{}); ok {

			if dept, ok := entityMap["department"].(string); !ok || dept == "" || strings.Contains(dept, "unspecified") {
				inferredDept := inferDepartment(input)
				if inferredDept != "" {
					entityMap["department"] = inferredDept
				}
			}
		}
	}

	return analysis
}

func isGenericAction(action string) bool {
	genericTerms := []string{"have", "want", "tell", "need", "you to"}
	for _, term := range genericTerms {
		if strings.Contains(action, term) {
			return true
		}
	}
	return false
}

func inferDepartment(input string) string {
	lowerInput := strings.ToLower(input)
	switch {
	case strings.Contains(lowerInput, "selling") || strings.Contains(lowerInput, "sales"):
		return "Sales"
	case strings.Contains(lowerInput, "market"):
		return "Marketing"
	case strings.Contains(lowerInput, "engineer") || strings.Contains(lowerInput, "technical"):
		return "Engineering"
	case strings.Contains(lowerInput, "finance") || strings.Contains(lowerInput, "accounting"):
		return "Finance"
	case strings.Contains(lowerInput, "customer"):
		return "Customer Support"
	default:
		return ""
	}
}

func extractJSON(text string) string {
	jsonStart := strings.Index(text, "{")
	jsonEnd := strings.LastIndex(text, "}")
	if jsonStart == -1 || jsonEnd == -1 {
		return ""
	}

	jsonStr := strings.TrimSpace(text[jsonStart : jsonEnd+1])
	jsonStr = strings.TrimPrefix(jsonStr, "```json")
	jsonStr = strings.TrimPrefix(jsonStr, "```")
	return strings.TrimSpace(jsonStr)
}

func displayResults(input string, analysis *AnalysisResult) {
	fmt.Println("\nChat-Bot:")

	// Instructions
	fmt.Println("Instructions:")
	if len(analysis.Instructions) > 0 {
		fmt.Printf("- %s\n", analysis.Instructions[0])
	} else {
		fmt.Println("- No clear instruction found")
	}

	// Actions
	fmt.Println("\nActions:")
	if len(analysis.Actions) > 0 {
		for _, action := range analysis.Actions {
			fmt.Printf("- %s\n", action)
		}
	} else {
		fmt.Println("- No actions identified")
	}

	if len(analysis.ResolvedEntities) > 0 {
		fmt.Println("\nContext:")
		for entity, details := range analysis.ResolvedEntities {
			if entityMap, ok := details.(map[string]interface{}); ok {
				name := getStringValue(entityMap, "name", entity)
				dept := getStringValue(entityMap, "department", "Department unspecified")
				fmt.Printf("- %s: %s\n", name, dept)
			}
		}
	}
	fmt.Println()

	logger.InfoLog.Printf("User: %s", input)
	logger.InfoLog.Printf("Instructions: %v", analysis.Instructions)
	logger.InfoLog.Printf("Actions: %v", analysis.Actions)
	if len(analysis.ResolvedEntities) > 0 {
		logger.InfoLog.Printf("Context: %v", analysis.ResolvedEntities)
	}
}

func getStringValue(m map[string]interface{}, key string, defaultValue string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return defaultValue
}

func formatTraits(traits []interface{}) string {
	var strTraits []string
	for _, t := range traits {
		if s, ok := t.(string); ok {
			strTraits = append(strTraits, s)
		}
	}
	return strings.Join(strTraits, ", ")
}
