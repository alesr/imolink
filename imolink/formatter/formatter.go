package formatter

import (
	"encoding/json"

	"encore.app/properties"
)

type propertyJSON struct {
	Reference      string   `json:"referencia"`
	PropertyType   string   `json:"tipo_imovel"`
	Name           string   `json:"nome"`
	Price          float64  `json:"preco"`
	Location       location `json:"localizacao"`
	Specifications specs    `json:"especificacoes"`
	YearBuilt      int      `json:"ano_construcao,omitempty"`
	Builder        *string  `json:"construtora,omitempty"`
	Features       []string `json:"caracteristicas,omitempty"`
	Description    *string  `json:"descricao,omitempty"`
}

type location struct {
	Street   string `json:"rua"`
	Number   int    `json:"numero"`
	District string `json:"bairro"`
	City     string `json:"cidade"`
	State    string `json:"estado"`
}

type specs struct {
	Area           float64 `json:"area"`
	NumBedrooms    int     `json:"quartos"`
	NumBathrooms   int     `json:"banheiros"`
	NumGarageSpots int     `json:"vagas_garagem"`
}

func FormatProperties(props []*properties.Property) string {
	properties := make([]propertyJSON, 0, len(props))

	for _, p := range props {
		prop := propertyJSON{
			Reference:    p.Reference,
			PropertyType: p.PropertyType,
			Name:         p.Name,
			Price:        p.Price,
			Location: location{
				Street:   p.Street,
				Number:   p.Number,
				District: p.District,
				City:     p.City,
				State:    p.State,
			},
			Specifications: specs{
				Area:           p.Area,
				NumBedrooms:    p.NumBedrooms,
				NumBathrooms:   p.NumBathrooms,
				NumGarageSpots: p.NumGarageSpots,
			},
			YearBuilt:   p.YearBuilt,
			Builder:     p.Builder,
			Features:    p.Features,
			Description: p.Description,
		}
		properties = append(properties, prop)
	}

	wrapper := struct {
		Properties []propertyJSON `json:"properties"`
	}{
		Properties: properties,
	}

	jsonData, err := json.MarshalIndent(wrapper, "", "  ")
	if err != nil {
		return ""
	}
	return string(jsonData)
}
