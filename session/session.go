package session

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"encore.app/internal/pkg/openaicli"
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
		currentRun, err := sm.openaiCli.GetRun(ctx, session.ThreadID, run.ID)
		if err != nil {
			return "", fmt.Errorf("could not get run status: %w", err)
		}

		switch currentRun.Status {
		case openaicli.RunStatusCompleted:
			goto COMPLETED
		case openaicli.RunStatusRequiresAction:
			if currentRun.RequiredAction == nil {
				return "", fmt.Errorf("invalid state: requires_action but no action specified")
			}
			if err := sm.handleFunctionCalling(ctx, session.ThreadID, currentRun); err != nil {
				return "", fmt.Errorf("could not handle function calling: %w", err)
			}
			time.Sleep(1 * time.Second)
		case openaicli.RunStatusFailed, openaicli.RunStatusCancelled, openaicli.RunStatusExpired:
			return "", fmt.Errorf("run failed with status: %s and error: %v", currentRun.Status, currentRun.LastError)
		}

		if currentRun.Status != openaicli.RunStatusCompleted {
			if err := sm.openaiCli.WaitForRun(ctx, session.ThreadID, run.ID); err != nil {
				if strings.Contains(err.Error(), "requires_action") {
					continue
				}
				return "", fmt.Errorf("could not wait for run: %w", err)
			}
		}
	}

COMPLETED:
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

func (sm *SessionManager) handleFunctionCalling(ctx context.Context, threadID string, run *openaicli.Run) error {
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

			if err := CreateLead(ctx, &CreateLeadInput{Name: args.Name}); err != nil {
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
		return session, nil
	}

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
