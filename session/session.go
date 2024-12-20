package session

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"encore.app/internal/pkg/openaicli"
)

type openaiCli interface {
	AddMessage(ctx context.Context, in openaicli.CreateMessageInput) error
	RunThread(ctx context.Context, threadID, assistantID string) (*openaicli.Run, error)
	WaitForRun(ctx context.Context, threadID, runID string) error
	GetMessages(ctx context.Context, threadID string) (*openaicli.ThreadMessageList, error)
	CreateThread(ctx context.Context) (*openaicli.Thread, error)
}

type SessionManager struct {
	mu        sync.RWMutex
	sessions  map[string]*Session // key: userID
	assistant *openaicli.Assistant
	openaiCli openaiCli
}

func NewSessionManager(assistant *openaicli.Assistant, openaiCli openaiCli) *SessionManager {
	return &SessionManager{
		sessions:  make(map[string]*Session),
		assistant: assistant,
		openaiCli: openaiCli,
	}
}

func (sm *SessionManager) SendMessage(ctx context.Context, userID, message string) (string, error) {
	session, err := sm.getOrCreateSession(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("could not get or create session: %w", err)
	}

	if err := sm.openaiCli.AddMessage(ctx, openaicli.CreateMessageInput{
		ThreadID: session.ThreadID,
		Message: openaicli.ThreadMessage{
			Role:    openaicli.RoleUser,
			Content: message,
		},
	}); err != nil {
		return "", fmt.Errorf("could not add message: %w", err)
	}

	run, err := sm.openaiCli.RunThread(ctx, session.ThreadID, sm.assistant.ID)
	if err != nil {
		return "", fmt.Errorf("could not run thread: %w", err)
	}

	for {
		err := sm.openaiCli.WaitForRun(ctx, session.ThreadID, run.ID)
		if err == nil {
			break // Run completed successfully
		}
		return "", fmt.Errorf("could not wait for run: %w", err)
	}

	// Get final response
	messages, err := sm.openaiCli.GetMessages(ctx, session.ThreadID)
	if err != nil {
		return "", fmt.Errorf("could not get messages: %w", err)
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
	thread, err := sm.openaiCli.CreateThread(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not create thread: %w", err)
	}

	sess := Session{
		ThreadID: thread.ID,
		UserID:   userID,
	}
	sm.sessions[userID] = &sess

	return &sess, nil
}
