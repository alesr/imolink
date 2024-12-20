package properties

import (
	"fmt"
	"time"
)

// Properties represents a list of real estate properties.
type Properties struct {
	Properties []*Property `json:"properties"`
}

// Property represents a real estate property.
type Property struct {
	ID                  string     `json:"id"`
	Name                string     `json:"name"`
	Area                float64    `json:"area"`
	NumBedrooms         int        `json:"numBedrooms"`
	NumBathrooms        int        `json:"numBathrooms"`
	NumGarageSpots      int        `json:"numGarageSpots"`
	Price               float64    `json:"price"`
	Street              string     `json:"street"`
	Number              int        `json:"number"`
	District            string     `json:"district"`
	City                string     `json:"city"`
	State               string     `json:"state"`
	PropertyType        string     `json:"propertyType"`
	Reference           string     `json:"reference"`
	Description         *string    `json:"description"`
	YearBuilt           int        `json:"yearBuilt"`
	Builder             *string    `json:"builder"`
	Features            []string   `json:"features"`
	PhotoBase64Data     *string    `json:"photoBase64Data,omitempty"`
	PhotoFormat         *string    `json:"photoFormat,omitempty"`
	PhotoUploadDate     *time.Time `json:"photoUploadDate,omitempty"`
	BlueprintBase64Data *string    `json:"blueprintBase64Data,omitempty"`
	BlueprintFormat     *string    `json:"blueprintFormat,omitempty"`
	BlueprintUploadDate *time.Time `json:"blueprintUploadDate,omitempty"`
	CreatedAt           time.Time  `json:"createdAt"`
	UpdatedAt           time.Time  `json:"updatedAt"`
}

// String returns a string representation of a property in Portuguese.
func (p *Property) String() string {
	description := "Não informado"
	if p.Description != nil {
		description = *p.Description
	}

	builder := "Não informado"
	if p.Builder != nil {
		builder = *p.Builder
	}

	features := "Nenhum"
	if len(p.Features) > 0 {
		features = fmt.Sprintf("%v", p.Features)
	}

	return fmt.Sprintf(
		"Nome: %s\nTipo: %s\nReferência: %s\nDescrição: %s\nÁrea: %.2f m²\n"+
			"Quartos: %d\nBanheiros: %d\nVagas: %d\nPreço: R$ %.2f\n"+
			"Endereço: %s, %d - %s, %s-%s\n"+
			"Ano de construção: %d\nConstrutora: %s\n"+
			"Características: %s\n"+
			"Criado em: %s\nAtualizado em: %s",
		p.Name, p.PropertyType, p.Reference, description, p.Area,
		p.NumBedrooms, p.NumBathrooms, p.NumGarageSpots, p.Price,
		p.Street, p.Number, p.District, p.City, p.State,
		p.YearBuilt, builder, features,
		p.CreatedAt.Format("02/01/2006 15:04:05"),
		p.UpdatedAt.Format("02/01/2006 15:04:05"),
	)
}

type ListInput struct {
	WithBase64Images bool `query:"with_base64_images"`
}
