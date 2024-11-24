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
	ID             string    `json:"-"`
	Name           string    `json:"name"`
	Area           float64   `json:"area"`
	NumBedrooms    int       `json:"numBedrooms"`
	NumBathrooms   int       `json:"numBathrooms"`
	NumGarageSpots int       `json:"numGarageSpots"`
	Price          float64   `json:"price"`
	Address        Address   `json:"address"`
	Type           string    `json:"type"`
	Info           Info      `json:"info"`
	CreatedAt      time.Time `json:"createdAt"`
	UpdatedAt      time.Time `json:"updatedAt"`
}

// String returns a string representation of a property in Portuguese.
func (p *Property) String() string {
	return fmt.Sprintf(
		"Nome: %s, Área: %.2f, Número de Quartos: %d, Número de Banheiros: %d, Número de Vagas de Garagem: %d, Preço: %.2f, Endereço: %s, Tipo: %s, Informações: %s",
		p.Name, p.Area, p.NumBedrooms, p.NumBathrooms, p.NumGarageSpots, p.Price, p.Address.String(), p.Type, p.Info.String(),
	)
}

// ImageData represents image data and metadata stored as base64
type ImageData struct {
	Format     string    `json:"format"`     // Image format/filename
	Base64Data string    `json:"data"`       // Base64 encoded image data
	UploadDate time.Time `json:"uploadDate"` // Upload timestamp
}

// Info represents additional information about a property
type Info struct {
	Reference   string    `json:"reference"`
	Photo       ImageData `json:"photo"`
	Blueprint   ImageData `json:"blueprint"`
	Description string    `json:"description"`
	YearBuilt   int       `json:"yearBuilt"`
	Builder     string    `json:"builder"`
	Features    []string  `json:"features"`
}

func (i *Info) String() string {
	return fmt.Sprintf(
		"Referência: %s, Foto: %s, Planta: %s, Descrição: %s, Ano de Construção: %d, Construtora: %s, Características: %s",
		i.Reference, i.Photo.Format, i.Blueprint.Format, i.Description, i.YearBuilt, i.Builder, i.Features,
	)
}

// Address represents a property address
type Address struct {
	Street   string `json:"street"`
	Number   int    `json:"number"`
	District string `json:"district"`
	City     string `json:"city"`
	State    string `json:"state"`
}

func (a *Address) String() string {
	return fmt.Sprintf(
		"Rua: %s, Número: %d, Bairro: %s, Cidade: %s, Estado: %s",
		a.Street, a.Number, a.District, a.City, a.State,
	)
}

// Question represents a question for the application
type Question struct {
	Question string `json:"question"`
}

// Response represents an application response
type Response struct {
	Message string `json:"message"`
}
