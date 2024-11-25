package app

import (
	"context"
	"embed"

	"encore.app/imolink"
	"encore.app/internal/pkg/apierror"
	"encore.app/properties"
	"encore.dev/beta/errs"
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

	if err := s.Delete(ctx); err != nil {
		return apierror.E("could not purge", err, errs.Internal)
	}

	sampleProps, err := getSampleProperties()
	if err != nil {
		return apierror.E("could not get sample properties", err, errs.Internal)
	}

	if err := properties.Create(
		ctx,
		&properties.Properties{Properties: sampleProps},
	); err != nil {
		return apierror.E("could not add sample properties", err, errs.Internal)
	}
	return nil
}

//encore:api auth method=DELETE path=/sample
func (s *Service) Delete(ctx context.Context) error {
	ctx = context.Background() // without cancellation
	if err := imolink.DeleteTrainingData(ctx); err != nil {
		return apierror.E("could not purge imolink", err, errs.Internal)
	}
	if err := properties.Delete(ctx); err != nil {
		return apierror.E("could not purge properties", err, errs.Internal)
	}
	return nil
}
