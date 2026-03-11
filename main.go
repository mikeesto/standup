package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"strings"
	"time"
)

const openRouterURL = "https://openrouter.ai/api/v1/chat/completions"

const model = "openai/gpt-oss-120b:free"

type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type ChatRequest struct {
	Model       string    `json:"model"`
	Messages    []Message `json:"messages"`
	Temperature float32   `json:"temperature"`
}

type ChatResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func getGitCommits() (string, error) {
	emailCmd := exec.Command("git", "config", "user.email")
	emailOut, err := emailCmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get git user email: %w", err)
	}
	email := strings.TrimSpace(string(emailOut))

	logCmd := exec.Command(
		"git", "log",
		"--since=24 hours ago",
		fmt.Sprintf("--author=%s", email),
		"--pretty=format:- %s (%h)",
	)
	logOut, err := logCmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("git error: %s", string(exitErr.Stderr))
		}
		return "", fmt.Errorf("failed to run git log: %w", err)
	}

	return strings.TrimSpace(string(logOut)), nil
}

func summarizeWithLLM(commits string, apiKey string) (string, error) {
	yesterday := time.Now().AddDate(0, 0, -1).Format("Monday, January 2")

	// System prompt defines the persona, rules, and examples
	systemPrompt := fmt.Sprintf(`You are a developer providing a daily standup update.
Your job is to read raw git commits and convert them into a concise, readable standup summary.

RULES:
1. Group related commits by feature, not just by module/directory.
2. KEEP specific feature names (e.g., 'uploaders', 'briefing/debriefing', 'ddpconfig').
3. DO NOT use generic phrases like "Added updates to" or "Updated changes in". Describe WHAT functionality was actually added or changed.
4. Start each bullet with a past-tense verb (Added, Fixed, Wired up, Implemented, Refactored).
5. Output plain text, formatted exactly as requested.

FORMAT:
Yesterday (%s):
- <summarized change 1>
- <summarized change 2>

EXAMPLE INPUT:
- fetch uploaders from ddpconfig in participantui (0c12497)
- wire up briefing/debriefing to participantui (91e3d1d)
- wire up uploader create/edit/delete in ddpconfig (ea1a222)

EXAMPLE OUTPUT:
Yesterday:
- Implemented uploader management (create, edit, delete, fetch) across ddpconfig and participant UI
- Wired up briefing and debriefing to the participant UI`, yesterday)

	reqBody := ChatRequest{
		Model:       model,
		Temperature: 0.2,
		Messages: []Message{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: fmt.Sprintf("Here are my commits from the last 24 hours:\n\n%s", commits)},
		},
	}

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", openRouterURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("HTTP-Referer", "https://github.com/mikeesto/standup")
	req.Header.Set("X-Title", "Standup CLI")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to call OpenRouter API: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var chatResp ChatResponse
	if err := json.Unmarshal(respBytes, &chatResp); err != nil {
		return "", fmt.Errorf("failed to parse response: %w\nRaw: %s", err, string(respBytes))
	}

	if chatResp.Error != nil {
		return "", fmt.Errorf("API error: %s", chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("no choices in response: %s", string(respBytes))
	}

	return strings.TrimSpace(chatResp.Choices[0].Message.Content), nil
}

func main() {
	apiKey := os.Getenv("OPENROUTER_MACBOOK_KEY")
	if apiKey == "" {
		fmt.Fprintln(os.Stderr, "Error: OPENROUTER_MACBOOK_KEY environment variable is not set")
		os.Exit(1)
	}

	fmt.Println("Standup summary (last 24 hours)")
	fmt.Println("--------------------------------")

	commits, err := getGitCommits()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting git commits: %v\n", err)
		os.Exit(1)
	}

	if commits == "" {
		fmt.Println("No commits found in the last 24 hours.")
		os.Exit(0)
	}

	fmt.Println("Raw commits:")
	fmt.Println(commits)
	fmt.Println()
	fmt.Println("Summarising with AI...")
	fmt.Println()

	summary, err := summarizeWithLLM(commits, apiKey)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error summarizing commits: %v\n", err)
		os.Exit(1)
	}

	fmt.Println(summary)
}
