package types

import (
	"encoding/json"
	"io"
)

type Model string

const (
	AssistantModel Model = "gpt-4o-mini"
	EmbeddingModel Model = "text-embedding-3-small"
)

const (
	RoleUser      = "user"
	RoleAssistant = "assistant"
	RoleSystem    = "system"
)

const (
	// Tool types
	ToolTypeFunction        = "function"
	ToolTypeCodeInterpreter = "code_interpreter"
	ToolTypeFileSearch      = "file_search"
)

// Common types
type Meta map[string]any

// Moving all shared types here
type Thread struct {
	ID        string `json:"id"`
	Object    string `json:"object"`
	CreatedAt int    `json:"created_at"`
	Metadata  Meta   `json:"metadata"`
}

type ThreadMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type CreateMessageInput struct {
	ThreadID string
	Message  ThreadMessage
}

type Run struct {
	ID             string          `json:"id"`
	Object         string          `json:"object"`
	CreatedAt      int64           `json:"created_at"`
	ThreadID       string          `json:"thread_id"`
	AssistantID    string          `json:"assistant_id"`
	Status         string          `json:"status"`
	StartedAt      int64           `json:"started_at,omitempty"`
	ExpiresAt      int64           `json:"expires_at,omitempty"`
	CancelledAt    int64           `json:"cancelled_at,omitempty"`
	FailedAt       int64           `json:"failed_at,omitempty"`
	CompletedAt    int64           `json:"completed_at,omitempty"`
	LastError      *Error          `json:"last_error,omitempty"`
	Model          string          `json:"model"`
	Instructions   string          `json:"instructions,omitempty"`
	Tools          []Tool          `json:"tools"`
	FileIDs        []string        `json:"file_ids"`
	RequiredAction *RequiredAction `json:"required_action,omitempty"`
}

type RequiredAction struct {
	Type      string     `json:"type"`
	ToolCalls []ToolCall `json:"tool_calls"`
}

type ToolCall struct {
	ID        string       `json:"id"`
	Type      string       `json:"type"`
	Function  FunctionCall `json:"function"`
	Arguments string       `json:"arguments"` // Add this line
}

type FunctionCall struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

type ToolOutput struct {
	ToolCallID string `json:"tool_call_id"`
	Output     string `json:"output"`
}

const (
	RunStatusQueued         = "queued"
	RunStatusInProgress     = "in_progress"
	RunStatusCompleted      = "completed"
	RunStatusFailed         = "failed"
	RunStatusCancelling     = "cancelling"
	RunStatusCancelled      = "cancelled"
	RunStatusExpired        = "expired"
	RunStatusRequiresAction = "requires_action"
)

type ThreadMessageList struct {
	Object  string           `json:"object"`
	Data    []MessageContent `json:"data"`
	FirstID string           `json:"first_id"`
	LastID  string           `json:"last_id"`
}

type MessageContent struct {
	ID        string    `json:"id"`
	Object    string    `json:"object"`
	CreatedAt int64     `json:"created_at"`
	ThreadID  string    `json:"thread_id"`
	Role      string    `json:"role"`
	Content   []Content `json:"content"`
}

type Function struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Parameters  json.RawMessage `json:"parameters"`
}

type AssistantCfg struct {
	Metadata      Meta          `json:"metadata,omitempty"`
	Name          string        `json:"name"`
	Description   string        `json:"description"`
	Model         Model         `json:"model"`
	Instructions  string        `json:"instructions"`
	Tools         []Tool        `json:"tools"`
	ToolResources ToolResources `json:"tool_resources,omitempty"`
}

type Tool struct {
	Type string `json:"type"`
}

type ToolResources struct {
	CodeInterpreter *CodeInterpreter `json:"code_interpreter,omitempty"`
	FileSearch      *FileSearch      `json:"file_search,omitempty"`
}

type CodeInterpreter struct {
	FileIDs []string `json:"file_ids"`
}

type FileSearch struct {
	VectorStoreIDs []string `json:"vector_store_ids"`
}

type Assistant struct {
	ID           string   `json:"id"`
	Object       string   `json:"object"`
	CreatedAt    int64    `json:"created_at"`
	Name         string   `json:"name"`
	Description  string   `json:"description"`
	Model        Model    `json:"model"`
	Instructions string   `json:"instructions"`
	Tools        []Tool   `json:"tools"`
	FileIDs      []string `json:"file_ids"`
	Metadata     Meta     `json:"metadata,omitempty"`
}

type Message struct {
	ID          string    `json:"id"`
	Object      string    `json:"object"`
	CreatedAt   int64     `json:"created_at"`
	ThreadID    string    `json:"thread_id"`
	Role        string    `json:"role"`
	Content     []Content `json:"content"`
	FileIDs     []string  `json:"file_ids"`
	AssistantID string    `json:"assistant_id,omitempty"`
	RunID       string    `json:"run_id,omitempty"`
	Metadata    Meta      `json:"metadata,omitempty"`
}

type Content struct {
	Type string    `json:"type"`
	Text TextValue `json:"text"`
}

type TextValue struct {
	Value       string       `json:"value"`
	Annotations []Annotation `json:"annotations,omitempty"`
	Citations   []Citation   `json:"citations,omitempty"`
}

type Annotation struct {
	Type         string `json:"type"`
	Text         string `json:"text"`
	FileCitation *struct {
		FileID string `json:"file_id"`
		Quote  string `json:"quote"`
	} `json:"file_citation,omitempty"`
}

type Citation struct {
	FileID string `json:"file_id"`
	Quote  string `json:"quote"`
}

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type BatchFileUpload struct {
	Data     []io.Reader
	Purpose  string
	Metadata Meta
}

type BatchUploadResult struct {
	FileIDs []string
	Errors  []error
}

type StoreVectorInput struct {
	FileID   string `json:"file_id"`
	Model    Model  `json:"model"`
	MaxChunk int    `json:"max_chunk,omitempty"`
}

type StoreVectorResponse struct {
	ID        string `json:"id"`
	Object    string `json:"object"`
	CreatedAt int64  `json:"created_at"`
	Status    string `json:"status"`
}

type ListResponse struct {
	Object  string `json:"object"`
	Data    []any  `json:"data"`
	FirstID string `json:"first_id"`
	LastID  string `json:"last_id"`
	HasMore bool   `json:"has_more"`
}

// FileUploadResponse represents the response from uploading a file.
type FileUploadResponse struct {
	ID        string `json:"id"`
	Object    string `json:"object"`
	Purpose   string `json:"purpose"`
	CreatedAt int64  `json:"created_at"`
}

// FileDetails represents the detailed information of a file.
type FileDetails struct {
	ID        string `json:"id"`
	Object    string `json:"object"`
	Purpose   string `json:"purpose"`
	CreatedAt int64  `json:"created_at"`
}

// AssistantFiles represents the files associated with an assistant.
type AssistantFiles struct {
	Object string          `json:"object"`
	Data   []AssistantFile `json:"data"`
}

type AssistantFile struct {
	ID        string `json:"id"`
	Object    string `json:"object"`
	CreatedAt int64  `json:"created_at"`
	FileID    string `json:"file_id"`
}

// Add these new types for run steps
type RunSteps struct {
	Object string    `json:"object"`
	Data   []RunStep `json:"data"`
}

type RunStep struct {
	ID          string      `json:"id"`
	Object      string      `json:"object"`
	CreatedAt   int64       `json:"created_at"`
	RunID       string      `json:"run_id"`
	Status      string      `json:"status"`
	StepDetails *StepDetail `json:"step_details"`
}

type StepDetail struct {
	Type      string     `json:"type"`
	ToolCalls []ToolCall `json:"tool_calls,omitempty"`
}
