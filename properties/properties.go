// Package properties provides a service to manage properties.
package properties

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"html/template"
	"net/http"
	"time"

	"encore.app/internal/pkg/apierror"

	"encore.dev/beta/errs"
	"encore.dev/storage/sqldb"
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

//encore:api public method=POST path=/properties
func (s *Service) Create(ctx context.Context, in *Properties) error {
	for _, prop := range in.Properties {
		exists, err := propertyExists(ctx, prop.ID)
		if err != nil {
			return fmt.Errorf("could not check property existence: %w", err)
		}

		if exists {
			if err := updateProperty(ctx, prop); err != nil {
				return fmt.Errorf("could not update property: %w", err)
			}
			continue
		}
		if err := insertProperty(ctx, prop); err != nil {
			return fmt.Errorf("could not store property: %w", err)
		}
	}
	return nil
}

//encore:api public method=GET path=/properties
func (s *Service) List(ctx context.Context, in ListInput) (*Properties, error) {
	query := `SELECT 
        id, name, area, num_bedrooms, num_bathrooms, num_garage_spots, 
        price, street, number, district, city, state, property_type,
        reference, description, year_built, builder, features`

	if in.WithBase64Images {
		query += `,
        photo_base64_data, photo_format, photo_upload_date,
        blueprint_base64_data, blueprint_format, blueprint_upload_date`
	}

	query += `,
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
		var p Property
		if in.WithBase64Images {
			if err := rows.Scan(
				&p.ID, &p.Name, &p.Area, &p.NumBedrooms, &p.NumBathrooms, &p.NumGarageSpots,
				&p.Price, &p.Street, &p.Number, &p.District, &p.City, &p.State, &p.PropertyType,
				&p.Reference, &p.Description, &p.YearBuilt, &p.Builder, &p.Features,
				&p.PhotoBase64Data, &p.PhotoFormat, &p.PhotoUploadDate,
				&p.BlueprintBase64Data, &p.BlueprintFormat, &p.BlueprintUploadDate,
				&p.CreatedAt, &p.UpdatedAt,
			); err != nil {
				return nil, apierror.E("could not scan properties", err, errs.Internal)
			}
		} else {
			if err := rows.Scan(
				&p.ID, &p.Name, &p.Area, &p.NumBedrooms, &p.NumBathrooms, &p.NumGarageSpots,
				&p.Price, &p.Street, &p.Number, &p.District, &p.City, &p.State, &p.PropertyType,
				&p.Reference, &p.Description, &p.YearBuilt, &p.Builder, &p.Features,
				&p.CreatedAt, &p.UpdatedAt,
			); err != nil {
				return nil, apierror.E("could not scan properties", err, errs.Internal)
			}
		}
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

//encore:api public method=DELETE path=/properties
func (s *Service) Delete(ctx context.Context) error {
	// Use DELETE instead of TRUNCATE since we don't have TRUNCATE permissions
	if _, err := db.Exec(ctx, `DELETE FROM properties`); err != nil {
		return fmt.Errorf("could not delete properties: %w", err)
	}
	return nil
}

func propertyExists(ctx context.Context, id string) (bool, error) {
	var exists bool
	err := db.QueryRow(ctx, `
		SELECT EXISTS(SELECT 1 FROM properties WHERE id = $1)
	`, id).Scan(&exists)

	if err != nil {
		return false, fmt.Errorf("error checking property existence: %w", err)
	}
	return exists, nil
}

func updateProperty(ctx context.Context, prop *Property) error {
	_, err := db.Exec(ctx, `
		UPDATE properties SET
			name = $1, area = $2, num_bedrooms = $3, num_bathrooms = $4,
			num_garage_spots = $5, price = $6, street = $7, number = $8,
			district = $9, city = $10, state = $11, property_type = $12,
			reference = $13, description = $14, year_built = $15,
			builder = $16, features = $17, updated_at = $18
		WHERE id = $19
	`,
		prop.Name, prop.Area, prop.NumBedrooms, prop.NumBathrooms,
		prop.NumGarageSpots, prop.Price, prop.Street, prop.Number,
		prop.District, prop.City, prop.State, prop.PropertyType,
		prop.Reference, prop.Description, prop.YearBuilt,
		prop.Builder, prop.Features, time.Now(), prop.ID,
	)
	return err
}

func insertProperty(ctx context.Context, prop *Property) error {
	now := time.Now()
	if _, err := db.Exec(ctx, `
		INSERT INTO properties (
			id, name, area, num_bedrooms, num_bathrooms,
			num_garage_spots, price, street, number,
			district, city, state, property_type,
			reference, description, year_built,
			builder, features, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
			$11, $12, $13, $14, $15, $16, $17, $18,
			$19, $20
		)
	`,
		prop.ID, prop.Name, prop.Area, prop.NumBedrooms, prop.NumBathrooms,
		prop.NumGarageSpots, prop.Price, prop.Street, prop.Number,
		prop.District, prop.City, prop.State, prop.PropertyType,
		prop.Reference, prop.Description, prop.YearBuilt,
		prop.Builder, prop.Features, now, now,
	); err != nil {
		return fmt.Errorf("could not insert property: %w", err)
	}
	return nil
}

func (s *Service) fetchProperty(ctx context.Context, ref string) (*Property, error) {
	query := `
        SELECT 
            id, name, area, num_bedrooms, num_bathrooms, num_garage_spots, 
            price, street, number, district, city, state, property_type,
            reference, description, year_built, builder, features,
            photo_base64_data, photo_format, photo_upload_date,
            blueprint_base64_data, blueprint_format, blueprint_upload_date,
            created_at, updated_at
        FROM properties
        WHERE reference = $1
        LIMIT 1`

	var p Property

	row := db.QueryRow(ctx, query, ref)
	if err := row.Scan(
		&p.ID, &p.Name, &p.Area, &p.NumBedrooms, &p.NumBathrooms, &p.NumGarageSpots,
		&p.Price, &p.Street, &p.Number, &p.District, &p.City, &p.State, &p.PropertyType,
		&p.Reference, &p.Description, &p.YearBuilt, &p.Builder, &p.Features,
		&p.PhotoBase64Data, &p.PhotoFormat, &p.PhotoUploadDate,
		&p.BlueprintBase64Data, &p.BlueprintFormat, &p.BlueprintUploadDate,
		&p.CreatedAt, &p.UpdatedAt,
	); err != nil {
		if errors.Is(err, sqldb.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("could not scan property: %w", err)
	}

	return &p, nil
}

func (s *Service) storeProperties(ctx context.Context, p *Property) error {
	query := `
        INSERT INTO properties (
            id, name, area, num_bedrooms, num_bathrooms, num_garage_spots, 
            price, street, number, district, city, state, property_type,
            reference, description, year_built, builder, features,
            photo_base64_data, photo_format, photo_upload_date,
            blueprint_base64_data, blueprint_format, blueprint_upload_date,
            created_at, updated_at
        ) VALUES (
            $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13,
            $14, $15, $16, $17, $18, $19, $20, $21, $22, $23, $24,
            $25, $26
        )`

	now := time.Now()

	if p.CreatedAt.IsZero() {
		p.CreatedAt = now
	}
	p.UpdatedAt = now

	if _, err := db.Exec(ctx, query,
		p.ID, p.Name, p.Area, p.NumBedrooms, p.NumBathrooms, p.NumGarageSpots,
		p.Price, p.Street, p.Number, p.District, p.City, p.State, p.PropertyType,
		p.Reference, p.Description, p.YearBuilt, p.Builder, p.Features,
		p.PhotoBase64Data, p.PhotoFormat, p.PhotoUploadDate,
		p.BlueprintBase64Data, p.BlueprintFormat, p.BlueprintUploadDate,
		p.CreatedAt, p.UpdatedAt,
	); err != nil {
		return fmt.Errorf("could not store property: %w", err)
	}
	return nil
}
