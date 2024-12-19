package openaicli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"encore.app/internal/pkg/openaicli/types"
)

func (c *Client) NewThread(ctx context.Context) (*types.Thread, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, fmt.Sprintf("%s/threads", c.baseURL), bytes.NewBuffer([]byte("{}")))
	if err != nil {
		return nil, fmt.Errorf("could not create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("OpenAI-Beta", "assistants=v2")

	resp, err := c.doWithRetry(req)
	if err != nil {
		return nil, fmt.Errorf("could not send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var thread types.Thread
	if err := json.NewDecoder(resp.Body).Decode(&thread); err != nil {
		return nil, fmt.Errorf("could not decode response: %w", err)
	}
	return &thread, nil
}

func (c *Client) AddMessage(ctx context.Context, in types.CreateMessageInput) error {
	jsonData, err := json.Marshal(in.Message)
	if err != nil {
		return fmt.Errorf("could not marshal message: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/threads/%s/messages", c.baseURL, in.ThreadID),
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return fmt.Errorf("could not create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("OpenAI-Beta", "assistants=v2")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("could not send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) GetMessages(ctx context.Context, threadID string) (*types.ThreadMessageList, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/threads/%s/messages", c.baseURL, threadID),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("could not create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("OpenAI-Beta", "assistants=v2")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	var messages types.ThreadMessageList
	if err := json.NewDecoder(resp.Body).Decode(&messages); err != nil {
		return nil, fmt.Errorf("could not decode response: %w", err)
	}
	return &messages, nil
}

func (c *Client) RunThread(ctx context.Context, threadID string, assistantID string) (*types.Run, error) {
	// Simplified input without tools
	input := struct {
		AssistantID string `json:"assistant_id"`
	}{
		AssistantID: assistantID,
	}

	jsonData, err := json.Marshal(input)
	if err != nil {
		return nil, fmt.Errorf("could not marshal run input: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/threads/%s/runs", c.baseURL, threadID),
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return nil, fmt.Errorf("could not create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("OpenAI-Beta", "assistants=v2")

	resp, err := c.doWithRetry(req)
	if err != nil {
		return nil, fmt.Errorf("could not send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		responseBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(responseBody))
	}

	var run types.Run
	if err := json.NewDecoder(resp.Body).Decode(&run); err != nil {
		return nil, fmt.Errorf("could not decode response: %w", err)
	}
	return &run, nil
}

// Add this new method to handle tool outputs
func (c *Client) SubmitToolOutputs(ctx context.Context, threadID string, runID string, outputs []types.ToolOutput) error {
	input := struct {
		ToolOutputs []types.ToolOutput `json:"tool_outputs"`
	}{
		ToolOutputs: outputs,
	}

	jsonData, err := json.Marshal(input)
	if err != nil {
		return fmt.Errorf("could not marshal tool outputs: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		fmt.Sprintf("%s/threads/%s/runs/%s/submit_tool_outputs", c.baseURL, threadID, runID),
		bytes.NewBuffer(jsonData),
	)
	if err != nil {
		return fmt.Errorf("could not create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("OpenAI-Beta", "assistants=v2")

	resp, err := c.doWithRetry(req)
	if err != nil {
		return fmt.Errorf("could not send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, body)
	}

	return nil
}

func (c *Client) GetRun(ctx context.Context, threadID, runID string) (*types.Run, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/threads/%s/runs/%s", c.baseURL, threadID, runID),
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

	var run types.Run
	if err := json.NewDecoder(resp.Body).Decode(&run); err != nil {
		return nil, fmt.Errorf("could not decode response: %w", err)
	}
	return &run, nil
}

func (c *Client) WaitForRun(ctx context.Context, threadID, runID string) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			run, err := c.GetRun(ctx, threadID, runID)
			if err != nil {
				return fmt.Errorf("failed to get run: %w", err)
			}

			switch run.Status {
			case types.RunStatusCompleted:
				return nil
			case types.RunStatusFailed:
				if run.LastError != nil {
					return fmt.Errorf("run failed: %s - %s", run.LastError.Code, run.LastError.Message)
				}
				return fmt.Errorf("run failed without error details")
			case types.RunStatusCancelled, types.RunStatusExpired:
				return fmt.Errorf("run ended with status: %s", run.Status)
			case types.RunStatusQueued, types.RunStatusInProgress:
				time.Sleep(time.Second)
				continue
			default:
				return fmt.Errorf("unknown run status: %s", run.Status)
			}
		}
	}
}
