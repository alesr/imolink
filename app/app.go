package app

import (
	"context"
	"embed"

	"encore.app/imolink"
	"encore.app/internal/pkg/apierror"
	"encore.app/properties"
)

var (
	//go:embed assets/*
	assetsFS embed.FS
)

//encore:service
type Service struct{}

func initService() (*Service, error) {
	return &Service{}, nil
}

//encore:api auth method=POST path=/sample
func (s *Service) Sample(ctx context.Context) error {
	ctx = context.Background() // without cancellation

	if err := s.Purge(ctx); err != nil {
		return apierror.E("could not purge", err)
	}

	sampleProps, err := getSampleProperties()
	if err != nil {
		return apierror.E("could not get sample properties", err)
	}

	if err := properties.AddProperties(
		ctx,
		&properties.Properties{Properties: sampleProps},
	); err != nil {
		return apierror.E("could not add sample properties", err)
	}
	return nil
}

//encore:api auth method=DELETE path=/purge
func (s *Service) Purge(ctx context.Context) error {
	ctx = context.Background() // without cancellation
	if err := imolink.Purge(ctx); err != nil {
		return apierror.E("could not purge imolink", err)
	}
	if err := properties.Purge(ctx); err != nil {
		return apierror.E("could not purge properties", err)
	}
	return nil
}
