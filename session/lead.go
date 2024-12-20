package session

import (
	"context"
	"fmt"
)

type CreateLeadInput struct {
	Name string
}

func CreateLead(ctx context.Context, input *CreateLeadInput) error {
	fmt.Printf("Creating lead: %s\n", input.Name)
	return nil
}
