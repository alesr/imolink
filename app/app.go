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

//encore:api auth method=POST path=/sample
func (s *Service) Sample(ctx context.Context) error {
	if err := s.Purge(ctx); err != nil {
		return fmt.Errorf("could not purge data: %w", err)
	}

	ctx = context.Background() // without cancellation

	if err := properties.AddProperties(
		ctx,
		&properties.Properties{Properties: getSampleProperties()},
	); err != nil {
		return fmt.Errorf("could not add sample properties: %w", err)
	}
	return nil
}

//encore:api auth method=DELETE path=/purge
func (s *Service) Purge(ctx context.Context) error {
	ctx = context.Background() // without cancellation
	if err := imolink.Purge(ctx); err != nil {
		return fmt.Errorf("could not purge imolink: %w", err)
	}
	if err := properties.Purge(ctx); err != nil {
		return fmt.Errorf("could not purge properties: %w", err)
	}
	return nil
}
