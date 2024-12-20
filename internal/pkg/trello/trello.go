package trello

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type TrelloCard struct {
	Name        string `json:"name"`
	Description string `json:"desc"`
	ListID      string `json:"idList"`
}

type TrelloAPI struct {
	APIKey  string
	Token   string
	BaseURL string
}

func NewTrelloAPI(apiKey, token string) *TrelloAPI {
	return &TrelloAPI{
		APIKey:  apiKey,
		Token:   token,
		BaseURL: "https://api.trello.com/1",
	}
}

func (t *TrelloAPI) CreateCard(card TrelloCard) error {
	url := fmt.Sprintf("%s/cards?idList=%s&key=%s&token=%s", t.BaseURL, card.ListID, t.APIKey, t.Token)

	payload, err := json.Marshal(card)
	if err != nil {
		return fmt.Errorf("error marshaling card data: %v", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(payload))
	if err != nil {
		return fmt.Errorf("error making POST request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}
	return nil
}
