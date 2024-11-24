package app

import (
	"context"
	"fmt"

	"encore.app/imolink"
	"encore.app/properties"
)

//encore:service
type Service struct{}

func initService() (*Service, error) {
	return &Service{}, nil
}

//encore:api private method=POST path=/purge
func (s *Service) Purge(ctx context.Context) error {
	if err := imolink.Purge(ctx); err != nil {
		return fmt.Errorf("could not purge imolink: %w", err)
	}
	if err := properties.Purge(ctx); err != nil {
		return fmt.Errorf("could not purge properties: %w", err)
	}
	return nil
}
