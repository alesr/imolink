package app

import (
	"context"
	"embed"

	"encore.app/domain"
	"encore.app/imolink"
	"encore.app/internal/pkg/apierror"
	"encore.app/internal/pkg/idutil"
	"encore.app/properties"
	"encore.dev/beta/errs"
	"encore.dev/rlog"
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

//encore:api public method=POST path=/sample
func (s *Service) Sample(ctx context.Context) error {
	if err := properties.Delete(ctx); err != nil {
		// If delete fails, wrap the error but continue since the table might be empty
		// Don't return here as we still want to try creating the sample data
		apierror.E("warning: could not purge existing data", err, errs.Internal)
	}

	sampleProps, err := getSampleProperties()
	if err != nil {
		return apierror.E("could not get sample properties", err, errs.Internal)
	}

	rlog.Debug("Generated %d sample properties", len(sampleProps))

	for _, prop := range sampleProps {
		if prop.ID == "" {
			id, err := idutil.NewID()
			if err != nil {
				return apierror.E("could not generate ID", err, errs.Internal)
			}
			prop.ID = id
		}
	}

	if err := properties.Create(ctx, &domain.Properties{Properties: sampleProps}); err != nil {
		return apierror.E("could not create sample properties", err, errs.Internal)
	}

	// Verify properties were created
	createdProps, err := properties.List(ctx, properties.ListInput{})
	if err != nil {
		return apierror.E("could not verify created properties", err, errs.Internal)
	}
	rlog.Debug("Created %d properties", len(createdProps.Properties))

	if err := imolink.InitializeAssistant(ctx); err != nil {
		return apierror.E("could not initialize assistant", err, errs.Internal)
	}
	return nil
}

//encore:api public method=DELETE path=/sample
func (s *Service) Purge(ctx context.Context) error {
	if err := properties.Delete(ctx); err != nil {
		return apierror.E("could not purge properties", err, errs.Internal)
	}
	return nil
}
