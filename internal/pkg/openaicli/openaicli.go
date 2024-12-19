package openaicli

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"time"

	"encore.app/internal/pkg/openaicli/types"
)

// Client represents an OpenAI API client
type Client struct {
	apiKey     string
	httpClient *http.Client
	baseURL    string
}

// ClientOption allows configuring the client
type ClientOption func(*Client)

// WithBaseURL sets a custom base URL for the client
func WithBaseURL(url string) ClientOption {
	return func(c *Client) {
		c.baseURL = url
	}
}

// New creates a new OpenAI client
func New(apiKey string, httpClient *http.Client, opts ...ClientOption) *Client {
	c := Client{
		apiKey:     apiKey,
		httpClient: httpClient,
		baseURL:    "https://api.openai.com/v1",
	}
	for _, opt := range opts {
		opt(&c)
	}
	return &c
}

// ListFiles retrieves a list of files that have been uploaded
func (c *Client) ListFiles(ctx context.Context) (*types.ListResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/files", nil)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.doWithRetry(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, body)
	}

	var fileList types.ListResponse
	if err := json.NewDecoder(resp.Body).Decode(&fileList); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}
	return &fileList, nil
}

// UploadFile uploads a file to OpenAI with enhanced logging
func (c *Client) UploadFile(ctx context.Context, data io.Reader, purpose string) (*types.FileUploadResponse, error) {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	filename := fmt.Sprintf("data_%d.json", time.Now().UnixNano())

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, fmt.Errorf("error creating form file: %w", err)
	}

	if _, err := io.Copy(part, data); err != nil {
		return nil, fmt.Errorf("error copying data to form file: %w", err)
	}

	if err := writer.WriteField("purpose", purpose); err != nil {
		return nil, fmt.Errorf("error writing purpose field: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("error closing multipart writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/files", &body)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.doWithRetry(req)
	if err != nil {
		return nil, fmt.Errorf("error sending request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		responseBody, _ := io.ReadAll(resp.Body)
		log.Printf("File upload failed. Status: %d, Response: %s", resp.StatusCode, responseBody)
		return nil, fmt.Errorf("API error (status %d): %s", resp.StatusCode, responseBody)
	}

	var uploadResp types.FileUploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&uploadResp); err != nil {
		return nil, fmt.Errorf("error decoding response: %w", err)
	}
	return &uploadResp, nil
}

func (c *Client) GetFileContent(ctx context.Context, fileID string) ([]byte, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/files/%s", c.baseURL, fileID),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.doWithRetry(req)
	if err != nil {
		return nil, fmt.Errorf("error retrieving file metadata: %w", err)
	}
	defer resp.Body.Close()

	var fileInfo types.FileDetails
	if err := json.NewDecoder(resp.Body).Decode(&fileInfo); err != nil {
		return nil, fmt.Errorf("error decoding file metadata: %w", err)
	}

	if fileInfo.Purpose == "assistants" {
		log.Printf("File %s is an assistant file and cannot be downloaded directly", fileID)
		return nil, fmt.Errorf("cannot download files with purpose: assistants")
	}

	contentReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		fmt.Sprintf("%s/files/%s/content", c.baseURL, fileID),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("error creating content request: %w", err)
	}

	contentReq.Header.Set("Authorization", "Bearer "+c.apiKey)

	contentResp, err := c.doWithRetry(contentReq)
	if err != nil {
		return nil, fmt.Errorf("error retrieving file content: %w", err)
	}
	defer contentResp.Body.Close()

	content, err := io.ReadAll(contentResp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %w", err)
	}
	return content, nil
}

type CreateVectorStoreRequest struct {
	Name        string   `json:"name"`
	Description string   `json:"description,omitempty"`
	FileIDs     []string `json:"file_ids"`
}

type VectorStoreResponse struct {
	ID         string `json:"id"`
	Object     string `json:"object"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	UsageBytes int    `json:"usage_bytes"`
	CreatedAt  int64  `json:"created_at"`
	FileCounts struct {
		InProgress int `json:"in_progress"`
		Completed  int `json:"completed"`
		Failed     int `json:"failed"`
		Cancelled  int `json:"cancelled"`
		Total      int `json:"total"`
	} `json:"file_counts"`
	Metadata     map[string]interface{} `json:"metadata"`
	ExpiresAfter interface{}            `json:"expires_after"`
	ExpiresAt    interface{}            `json:"expires_at"`
	LastActiveAt int64                  `json:"last_active_at"`
}

func (c *Client) CreateVectorStore(ctx context.Context, in *CreateVectorStoreRequest) (*VectorStoreResponse, error) {
	body, err := json.Marshal(in)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	req, err := http.NewRequest(http.MethodPost, c.baseURL+"/vector_stores", bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("OpenAI-Beta", "assistants=v2")

	resp, err := c.doWithRetry(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		b, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response body: %w", err)
		}
		return nil, fmt.Errorf("failed to create vector store, status: '%s', body: '%s'", resp.Status, b)
	}

	var out VectorStoreResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}
	return &out, nil
}

func (c *Client) WaitForVectorStoreCompletion(ctx context.Context, vectorStoreID string, timeout, maxDelay time.Duration) error {
	startTime := time.Now()
	delay := 1 * time.Second // initial delay for exponential backoff

	for {
		req, err := http.NewRequest(http.MethodGet, c.baseURL+"/vector_stores/"+vectorStoreID, nil)
		if err != nil {
			return fmt.Errorf("failed to create HTTP request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+c.apiKey)

		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
		req.Header.Set("OpenAI-Beta", "assistants=v2")

		resp, err := c.doWithRetry(req)
		if err != nil {
			return fmt.Errorf("failed to send HTTP request: %w", err)
		}
		defer resp.Body.Close()

		var response VectorStoreResponse
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}

		fmt.Printf("Vector store response: %+v\n", response)

		if response.Status == "completed" {
			fmt.Println("Vector store creation completed successfully.")
			return nil
		}

		if response.Status == "failed" {
			return fmt.Errorf("vector store creation failed")
		}

		if time.Since(startTime) > timeout {
			return fmt.Errorf("timeout reached while waiting for vector store completion")
		}

		if delay < maxDelay {
			delay *= 2 // Double the delay for the next attempt
		}

		fmt.Printf("Waiting for %v before retrying...\n", delay)
		time.Sleep(delay)
	}
}
