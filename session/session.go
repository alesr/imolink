package session

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"encore.app/internal/pkg/openaicli"
	"encore.app/internal/pkg/trello"
	"encore.app/leads"
	"encore.dev/storage/sqldb"
)

const (
	cleanupInterval = 1 * time.Hour
	sessionTimeout  = 24 * time.Hour
	roleAssistant   = "assistant"
	messageTextType = "text"
)

type openaiCli interface {
	AddMessage(ctx context.Context, in openaicli.CreateMessageInput) error
	RunThread(ctx context.Context, threadID, assistantID string) (*openaicli.Run, error)
	WaitForRun(ctx context.Context, threadID, runID string) error
	GetMessages(ctx context.Context, threadID string) (*openaicli.ThreadMessageList, error)
	CreateThread(ctx context.Context) (*openaicli.Thread, error)
	GetRun(ctx context.Context, threadID, runID string) (*openaicli.Run, error)
	SubmitToolOutputs(ctx context.Context, threadID, runID string, outputs []openaicli.ToolOutput) error
	GetRunSteps(ctx context.Context, threadID, runID string) (*openaicli.RunSteps, error)
}

type SessionManager struct {
	mu              sync.RWMutex
	sessions        map[string]*Session
	assistant       *openaicli.Assistant
	openaiCli       openaiCli
	cleanupInterval time.Duration
	sessionTimeout  time.Duration
}

func NewSessionManager(assistant *openaicli.Assistant, openaiCli openaiCli) *SessionManager {
	sm := &SessionManager{
		sessions:        make(map[string]*Session),
		assistant:       assistant,
		openaiCli:       openaiCli,
		cleanupInterval: 1 * time.Hour,
		sessionTimeout:  24 * time.Hour,
	}

	go sm.cleanupLoop()
	return sm
}

func (sm *SessionManager) cleanupLoop() {
	ticker := time.NewTicker(sm.cleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		sm.cleanup()
	}
}

func (sm *SessionManager) cleanup() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	threshold := time.Now().Add(-sm.sessionTimeout)
	for userID, session := range sm.sessions {
		if session.LastAccessedAt.Before(threshold) {
			delete(sm.sessions, userID)
		}
	}
}

func (sm *SessionManager) SendMessage(ctx context.Context, db *sqldb.Database, trelloAPI *trello.TrelloAPI, userID, message string) (string, error) {
	session, err := sm.getOrCreateSession(ctx, userID)
	if err != nil {
		return "", fmt.Errorf("could not get or create session: %w", err)
	}

	// If name is collected, append this context to the message
	if session.NameCollected {
		message = fmt.Sprintf("(Context: User's name is %s) %s", session.CollectedName, message)
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

	if err := sm.processRun(ctx, db, trelloAPI, session.ThreadID, run.ID); err != nil {
		return "", err
	}
	return sm.getAssistantResponse(ctx, session.ThreadID)
}

func (sm *SessionManager) processRun(ctx context.Context, db *sqldb.Database, trelloAPI *trello.TrelloAPI, threadID, runID string) error {
	for {
		currentRun, err := sm.openaiCli.GetRun(ctx, threadID, runID)
		if err != nil {
			return fmt.Errorf("could not get run status: %w", err)
		}

		switch currentRun.Status {
		case openaicli.RunStatusCompleted:
			return nil
		case openaicli.RunStatusRequiresAction:
			if currentRun.RequiredAction == nil {
				return fmt.Errorf("invalid state: requires_action but no action specified")
			}
			if err := sm.handleFunctionCalling(ctx, db, trelloAPI, threadID, currentRun); err != nil {
				return fmt.Errorf("could not handle function calling: %w", err)
			}
			time.Sleep(1 * time.Second)
		case openaicli.RunStatusFailed, openaicli.RunStatusCancelled, openaicli.RunStatusExpired:
			return fmt.Errorf("run failed with status: %s and error: %v", currentRun.Status, currentRun.LastError)
		}

		if currentRun.Status != openaicli.RunStatusCompleted {
			if err := sm.openaiCli.WaitForRun(ctx, threadID, runID); err != nil {
				if strings.Contains(err.Error(), "requires_action") {
					continue
				}
				return fmt.Errorf("could not wait for run: %w", err)
			}
		}
	}
}

func (sm *SessionManager) getAssistantResponse(ctx context.Context, threadID string) (string, error) {
	messages, err := sm.openaiCli.GetMessages(ctx, threadID)
	if err != nil {
		return "", fmt.Errorf("could not get messages: %w", err)
	}

	if len(messages.Data) == 0 {
		return "", fmt.Errorf("no messages returned")
	}

	var finalResponse strings.Builder
	mostRecentMsg := messages.Data[0]
	if mostRecentMsg.Role == roleAssistant && len(mostRecentMsg.Content) > 0 {
		for _, content := range mostRecentMsg.Content {
			if content.Type == messageTextType {
				finalResponse.WriteString(content.Text.Value)
				finalResponse.WriteString("\n")
			}
		}
	}
	return finalResponse.String(), nil
}

func (sm *SessionManager) handleFunctionCalling(ctx context.Context, db *sqldb.Database, trelloAPI *trello.TrelloAPI, threadID string, run *openaicli.Run) error {
	if run.RequiredAction == nil {
		return nil
	}

	steps, err := sm.openaiCli.GetRunSteps(ctx, threadID, run.ID)
	if err != nil {
		return fmt.Errorf("could not get run steps: %w", err)
	}

	var toolCalls []openaicli.ToolCall
	if len(run.RequiredAction.ToolCalls) > 0 {
		toolCalls = run.RequiredAction.ToolCalls
	} else if len(steps.Data) > 0 && steps.Data[0].StepDetails != nil {
		toolCalls = steps.Data[0].StepDetails.ToolCalls
	}

	if len(toolCalls) == 0 {
		return nil
	}

	var toolOutputs []openaicli.ToolOutput

	for _, toolCall := range toolCalls {
		if toolCall.Type != openaicli.ToolTypeFunction {
			continue
		}

		switch toolCall.Function.Name {
		case "lead":
			var args struct {
				Name string `json:"name"`
			}
			if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &args); err != nil {
				return fmt.Errorf("could not parse lead arguments: %w", err)
			}

			// Find session by threadID
			sm.mu.Lock()
			var userPhone string
			var session *Session
			for _, sess := range sm.sessions {
				if sess.ThreadID == threadID {
					userPhone = sess.UserID // UserID contains the WhatsApp number
					session = sess
					break
				}
			}
			sm.mu.Unlock()

			if session == nil {
				return fmt.Errorf("no session found for thread %s", threadID)
			}

			// Update session with name information
			session.NameCollected = true
			session.CollectedName = args.Name

			if userPhone == "" {
				return fmt.Errorf("no session found for thread %s", threadID)
			}

			if err := leads.CreateLead(ctx, db, trelloAPI, &leads.CreateLeadInput{
				Name: args.Name,
				// Clean up the phone number by removing the WhatsApp suffix.
				Phone: strings.Split(strings.Split(userPhone, "@")[0], ":")[0],
			}); err != nil {
				return fmt.Errorf("could not create lead: %w", err)
			}

			toolOutputs = append(toolOutputs, openaicli.ToolOutput{
				ToolCallID: toolCall.ID,
				Output:     "Lead created successfully",
			})
		}
	}

	if len(toolOutputs) > 0 {
		if err := sm.openaiCli.SubmitToolOutputs(ctx, threadID, run.ID, toolOutputs); err != nil {
			return fmt.Errorf("could not submit tool outputs: %w", err)
		}
	}
	return nil
}

func (sm *SessionManager) getOrCreateSession(ctx context.Context, userID string) (*Session, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	if session, exists := sm.sessions[userID]; exists {
		session.LastAccessedAt = time.Now()
		return session, nil
	}

	thread, err := sm.openaiCli.CreateThread(ctx)
	if err != nil {
		return nil, fmt.Errorf("could not create thread: %w", err)
	}

	sess := Session{
		ThreadID:       thread.ID,
		UserID:         userID,
		LastAccessedAt: time.Now(),
	}
	sm.sessions[userID] = &sess

	return &sess, nil
}
