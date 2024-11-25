package openaicli

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"

	"encore.app/internal/pkg/openaicli/types"
)

type Session struct {
	ThreadID string
	UserID   string
}

type SessionManager struct {
	mu        sync.RWMutex
	sessions  map[string]*Session // key: userID
	assistant *types.Assistant
	openaiCli *Client
}

func NewSessionManager(assistant *types.Assistant, openaiCli *Client) *SessionManager {
	return &SessionManager{
		sessions:  make(map[string]*Session),
		assistant: assistant,
		openaiCli: openaiCli,
	}
}

func (sm *SessionManager) SendMessage(ctx context.Context, userID, message string) (string, error) {
	session, err := sm.getOrCreateSession(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("failed to get session: %w", err)
	}

	if err := sm.openaiCli.AddMessage(ctx, types.CreateMessageInput{
		ThreadID: session.ThreadID,
		Message: types.ThreadMessage{
			Role:    types.RoleUser,
			Content: message,
		},
	}); err != nil {
		return "", fmt.Errorf("failed to add message: %w", err)
	}

	run, err := sm.openaiCli.RunThread(ctx, session.ThreadID, sm.assistant.ID)
	if err != nil {
		return "", fmt.Errorf("failed to run thread: %w", err)
	}

	for {
		err := sm.openaiCli.WaitForRun(ctx, session.ThreadID, run.ID)
		if err == nil {
			break // Run completed successfully
		}
		return "", fmt.Errorf("failed to wait for run: %w", err)
	}

	// Get final response
	messages, err := sm.openaiCli.GetMessages(ctx, session.ThreadID)
	if err != nil {
		return "", fmt.Errorf("failed to get messages: %w", err)
	}

	var finalResponse strings.Builder
	if len(messages.Data) == 0 {
		return "", fmt.Errorf("no messages returned")
	}

	mostRecentMsg := messages.Data[0]

	if mostRecentMsg.Role == "assistant" && len(mostRecentMsg.Content) > 0 {
		for _, content := range mostRecentMsg.Content {
			if content.Type == "text" {
				finalResponse.WriteString(content.Text.Value)
				finalResponse.WriteString("\n")
			}
		}
	}
	return finalResponse.String(), nil
}

func (sm *SessionManager) getOrCreateSession(ctx context.Context, userID string) (*Session, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if session, exists := sm.sessions[userID]; exists {
		return session, nil
	}

	// Create thread without system message first
	thread, err := sm.openaiCli.NewThread(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create thread: %w", err)
	}

	session := Session{
		ThreadID: thread.ID,
		UserID:   userID,
	}
	sm.sessions[userID] = &session

	return &session, nil
}

// Add this new method to get run steps
func (c *Client) GetRunSteps(ctx context.Context, threadID, runID string) (*types.RunSteps, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/threads/%s/runs/%s/steps", c.baseURL, threadID, runID),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("could not create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("OpenAI-Beta", "assistants=v2")

	resp, err := c.doWithRetry(req)
	if err != nil {
		return nil, fmt.Errorf("could not send request: %w", err)
	}
	defer resp.Body.Close()

	var steps types.RunSteps
	if err := json.NewDecoder(resp.Body).Decode(&steps); err != nil {
		return nil, fmt.Errorf("could not decode response: %w", err)
	}

	return &steps, nil
}
