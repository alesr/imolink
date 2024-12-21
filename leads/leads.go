package leads

import (
	"context"
	"fmt"
	"os"

	"encore.app/internal/pkg/idutil"
	"encore.app/internal/pkg/trello"
	"encore.dev/storage/sqldb"
)

const newLeadsTrelloLane = "6765c8d942977be5554e82d8"

type CreateLeadInput struct {
	Name  string
	Phone string
}

func CreateLead(ctx context.Context, db *sqldb.Database, trelloAPI *trello.TrelloAPI, input *CreateLeadInput) error {
	id, err := idutil.NewID()
	if err != nil {
		return fmt.Errorf("could not generate ID: %w", err)
	}

	go func() {
		if _, err := db.Exec(context.Background(), `
			INSERT INTO leads (id, name, phone)
			VALUES ($1, $2, $3)
		`, id, input.Name, input.Phone); err != nil {
			fmt.Fprintf(os.Stderr, "could not insert lead: %v\n", err)
			return
		}

		if err := trelloAPI.CreateCard(trello.TrelloCard{
			Name:        input.Name,
			Description: fmt.Sprintf("Phone: %s", input.Phone),
			ListID:      newLeadsTrelloLane,
		}); err != nil {
			fmt.Fprintf(os.Stderr, "could not create Trello card: %v\n", err)
			return
		}

		fmt.Printf("Created lead - ID: %s, Name: %s, Phone: %s\n", id, input.Name, input.Phone)
	}()
	return nil
}
