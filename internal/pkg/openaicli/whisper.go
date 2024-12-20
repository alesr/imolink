package openaicli

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

const whisperModel = "whisper-1"

// TranscribeAudio transcribes the audio from the given input.
func (c *Client) TranscribeAudio(in TranscribeAudioInput) ([]byte, error) {
	body := bytes.Buffer{}
	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile("file", in.Name)
	if err != nil {
		return nil, fmt.Errorf("could not create form file: %w", err)
	}

	if _, err := io.Copy(part, in.Data); err != nil {
		return nil, fmt.Errorf("could not copy data: %w", err)
	}

	if err := writer.WriteField("model", whisperModel); err != nil {
		return nil, fmt.Errorf("could not write model field: %w", err)
	}

	if err := writer.WriteField("response_format", "text"); err != nil {
		return nil, fmt.Errorf("could not write response_format field: %w", err)
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("could not close writer: %w", err)
	}

	request, err := http.NewRequest(http.MethodPost, c.baseURL+"/audio/transcriptions", &body)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	request.Header.Set("Authorization", "Bearer "+c.apiKey)
	request.Header.Set("Content-Type", writer.FormDataContentType())

	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("sending request: %w", err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(response.Body)
		return nil, fmt.Errorf("api error (status %d): %s", response.StatusCode, string(respBody))
	}

	b, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}
	return b, nil
}
