package openaicli

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCreateChatCompletition(t *testing.T) {
	t.Parallel()

	t.Run("just testing if the payload is well formed", func(t *testing.T) {
		t.Parallel()
		server := httptest.NewServer(nil)
		defer server.Close()

		client := Client{
			apiKey:     "fake-api-key",
			httpClient: server.Client(),
		}

		_, err := client.CreateChatCompletition(context.TODO(), CompletitionRequest{
			Messages: []Message{
				{
					Role:    "user",
					Content: "What is the meaning of life?",
				},
			},
		})
		require.Error(t, err)
	})
}
