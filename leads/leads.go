package leads

import (
	"context"
	"fmt"

	"encore.app/internal/pkg/idutil"
	"encore.dev/storage/sqldb"
)

var db = sqldb.NewDatabase("leads", sqldb.DatabaseConfig{
	Migrations: "./migrations",
})

type CreateLeadInput struct {
	Name  string
	Phone string
}

//encore:service
type Service struct{}

func CreateLead(ctx context.Context, input *CreateLeadInput) error {
	id, err := idutil.NewID()
	if err != nil {
		return fmt.Errorf("could not generate ID: %w", err)
	}

	if _, err = db.Exec(ctx, `
        INSERT INTO leads (id, name, phone)
        VALUES ($1, $2, $3)
    `, id, input.Name, input.Phone); err != nil {
		return fmt.Errorf("failed to insert lead: %w", err)
	}

	fmt.Printf("Created lead - ID: %s, Name: %s, Phone: %s\n", id, input.Name, input.Phone)
	return nil
}
