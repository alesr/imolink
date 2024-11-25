package properties

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"encore.app/imolink"
	"encore.app/internal/pkg/apierror"
	"encore.app/internal/pkg/idutil"

	"encore.dev/beta/errs"
	"encore.dev/storage/sqldb"
	"golang.org/x/sync/errgroup"
)

var (
	db = sqldb.NewDatabase("properties", sqldb.DatabaseConfig{
		Migrations: "./migrations",
	})

	//go:embed templates/*.html
	templatesFS embed.FS

	secrets struct {
		OpenAIKey   string
		BearerToken string
	}
)

//encore:service
type Service struct {
	templ *template.Template
}

func initService() (*Service, error) {
	tmpl, err := template.ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		return nil, fmt.Errorf("could not parse templates: %w", err)
	}
	return &Service{templ: tmpl}, nil
}

//encore:api auth method=POST path=/properties
func (s *Service) Create(ctx context.Context, in *Properties) error {
	group, _ := errgroup.WithContext(context.Background())

	for _, p := range in.Properties {
		group.Go(func() error {
			id, err := idutil.NewID()
			if err != nil {
				return fmt.Errorf("could not generate ID: %w", err)
			}
			p.ID = id

			if err := s.storeProperties(ctx, p); err != nil {
				return fmt.Errorf("could not store properties: %w", err)
			}

			p.Info.Photo = ImageData{}
			p.Info.Blueprint = ImageData{}

			imolink.NewPropertiesTopic.Publish(
				ctx,
				&imolink.NewPropertyEvent{
					Data: p.String(),
				},
			)
			return nil
		})
	}

	if err := group.Wait(); err != nil {
		return apierror.E("could not add properties", err, errs.Internal)
	}
	return nil
}

//encore:api auth method=GET path=/properties
func (s *Service) Get(ctx context.Context) (*Properties, error) {
	query := `
		SELECT
			id, name, area, num_bedrooms, num_bathrooms, num_garage_spots,
			price, street, number, district, city, state, property_type,
			reference, 
			photo_base64_data, photo_format, photo_upload_date,
			blueprint_base64_data, blueprint_format, blueprint_upload_date,
			description, year_built, builder,
			features,
			created_at, updated_at
		FROM properties`

	rows, err := db.Query(ctx, query)
	if err != nil {
		return nil, apierror.E("could not fetch properties", err, errs.Internal)
	}
	defer rows.Close()

	props := Properties{
		Properties: make([]*Property, 0),
	}

	for rows.Next() {
		var (
			p    Property
			addr Address
			info Info
		)

		if err := rows.Scan(
			&p.ID, &p.Name, &p.Area, &p.NumBedrooms, &p.NumBathrooms, &p.NumGarageSpots,
			&p.Price, &addr.Street, &addr.Number, &addr.District, &addr.City, &addr.State,
			&p.Type, &info.Reference,
			&info.Photo.Base64Data, &info.Photo.Format, &info.Photo.UploadDate,
			&info.Blueprint.Base64Data, &info.Blueprint.Format, &info.Blueprint.UploadDate,
			&info.Description, &info.YearBuilt, &info.Builder, &info.Features,
			&p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			if errors.Is(err, sqldb.ErrNoRows) {
				return nil, nil
			}
			return nil, apierror.E("could not scan properties", err, errs.Internal)
		}

		p.Address = addr
		p.Info = info
		props.Properties = append(props.Properties, &p)
	}
	return &props, nil
}

//encore:api public raw path=/properties/:ref
func (s *Service) Serve(w http.ResponseWriter, req *http.Request) {
	ref := req.URL.Path[len("/properties/"):]

	prop, err := s.fetchProperty(req.Context(), ref)
	if err != nil {
		http.Error(w, fmt.Sprintf("could not fetch property: %v", err), http.StatusInternalServerError)
		return
	}

	if prop == nil {
		http.Error(w, "property not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/html")
	if err := s.templ.ExecuteTemplate(w, "property", prop); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (s *Service) fetchProperty(ctx context.Context, ref string) (*Property, error) {
	query := `
        SELECT 
            id, name, area, num_bedrooms, num_bathrooms, num_garage_spots, 
            price, street, number, district, city, state, property_type,
            reference, 
            photo_base64_data, photo_format, photo_upload_date,
            blueprint_base64_data, blueprint_format, blueprint_upload_date,
            COALESCE(description, ''), year_built, COALESCE(builder, ''), 
            COALESCE(features, ARRAY[]::text[])
        FROM properties
        WHERE reference = $1
        LIMIT 1`

	var (
		p    Property
		addr Address
		info Info
	)

	row := db.QueryRow(ctx, query, ref)
	if err := row.Scan(
		&p.ID, &p.Name, &p.Area, &p.NumBedrooms, &p.NumBathrooms, &p.NumGarageSpots,
		&p.Price, &addr.Street, &addr.Number, &addr.District, &addr.City, &addr.State,
		&p.Type, &info.Reference,
		&info.Photo.Base64Data, &info.Photo.Format, &info.Photo.UploadDate,
		&info.Blueprint.Base64Data, &info.Blueprint.Format, &info.Blueprint.UploadDate,
		&info.Description, &info.YearBuilt, &info.Builder, &info.Features,
	); err != nil {
		if errors.Is(err, sqldb.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("could not scan property: %w", err)
	}

	p.Address = addr
	p.Info = info

	return &p, nil
}

func (s *Service) storeProperties(ctx context.Context, p *Property) error {
	query := `
        INSERT INTO properties (
            id, name, area, num_bedrooms, num_bathrooms, num_garage_spots, 
            price, street, number, district, city, state, property_type,
            reference, 
            photo_base64_data, photo_format, photo_upload_date,
            blueprint_base64_data, blueprint_format, blueprint_upload_date,
            description, year_built, builder, 
            features,
            created_at, updated_at
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13,
            $14, $15, $16, $17, $18, $19, $20, $21, $22, $23,
            $24, $25, $26
        )`

	now := time.Now()

	if p.CreatedAt.IsZero() {
		p.CreatedAt = now
	}
	p.UpdatedAt = now

	if _, err := db.Exec(ctx, query,
		p.ID, p.Name, p.Area, p.NumBedrooms, p.NumBathrooms, p.NumGarageSpots,
		p.Price, p.Address.Street, p.Address.Number, p.Address.District,
		p.Address.City, p.Address.State, p.Type,
		p.Info.Reference,
		p.Info.Photo.Base64Data, p.Info.Photo.Format, p.Info.Photo.UploadDate,
		p.Info.Blueprint.Base64Data, p.Info.Blueprint.Format, p.Info.Blueprint.UploadDate,
		p.Info.Description, p.Info.YearBuilt, p.Info.Builder, p.Info.Features,
		p.CreatedAt, p.UpdatedAt,
	); err != nil {
		return fmt.Errorf("could not store property: %w", err)
	}
	return nil
}

//encore:api private method=DELETE path=/properties
func (s *Service) Delete(ctx context.Context) error {
	if _, err := db.Exec(ctx, "DELETE FROM properties"); err != nil {
		return fmt.Errorf("could not purge properties: %w", err)
	}
	return nil
}
