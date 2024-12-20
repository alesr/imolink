package imolink

import (
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"math/rand"

	"encore.app/properties"
)

func strPtr(s string) *string {
	return &s
}

type propertyTemplate struct {
	name        string
	area        float64
	bedrooms    int
	bathrooms   int
	garageSpots int
	price       float64
	street      string
	number      int
	district    string
	propType    string
	reference   string
	yearBuilt   int
	builder     string
	features    []string
	description string
}

func getSampleProperties() ([]*properties.Property, error) {
	now := time.Now()
	templates := []propertyTemplate{
		{
			name:        "Mansão Jardins",
			area:        850.50,
			bedrooms:    5,
			bathrooms:   6,
			garageSpots: 4,
			price:       3850000.00,
			street:      "Rua José Steremberg",
			number:      235,
			district:    "Jardins",
			propType:    "casa",
			reference:   "REF123",
			yearBuilt:   2018,
			builder:     "Construtora Celi",
			features:    []string{"piscina", "sauna", "quadra de tênis", "churrasqueira"},
			description: "Mansão luxuosa no bairro mais nobre de Aracaju, com vista privilegiada e acabamento de alto padrão",
		},
		{
			name:        "Edifício Le Jardin",
			area:        245.75,
			bedrooms:    4,
			bathrooms:   3,
			garageSpots: 3,
			price:       1250000.00,
			street:      "Avenida Ministro Geraldo Barreto Sobral",
			number:      1578,
			district:    "Grageru",
			propType:    "apartamento",
			reference:   "REF345",
			yearBuilt:   2020,
			builder:     "Cosil Construções",
			features:    []string{"piscina", "ginásio", "churrasqueira", "varanda"},
			description: "Apartamento de alto padrão no Grageru, próximo aos principais shoppings da cidade",
		},
		{
			name:        "Residencial Atalaia Sul",
			area:        320.30,
			bedrooms:    3,
			bathrooms:   4,
			garageSpots: 2,
			price:       890000.00,
			street:      "Avenida Santos Dumont",
			number:      963,
			district:    "Atalaia",
			propType:    "sobrado",
			reference:   "REF678",
			yearBuilt:   2019,
			builder:     "Habitacional Construções",
			features:    []string{"piscina", "churrasqueira", "varanda"},
			description: "Sobrado moderno próximo à praia de Atalaia, com vista para o mar",
		},
		{
			name:        "Condomínio Farol da Ilha",
			area:        178.45,
			bedrooms:    3,
			bathrooms:   2,
			garageSpots: 2,
			price:       650000.00,
			street:      "Rua Niceu Dantas",
			number:      451,
			district:    "Farolândia",
			propType:    "apartamento",
			reference:   "REF901",
			yearBuilt:   2021,
			builder:     "União Engenharia",
			features:    []string{"piscina", "ginásio", "varanda"},
			description: "Apartamento familiar em área universitária, próximo à UNIT e principais supermercados",
		},
		{
			name:        "Palácio de Santana",
			area:        620.80,
			bedrooms:    4,
			bathrooms:   5,
			garageSpots: 3,
			price:       2750000.00,
			street:      "Praça Fausto Cardoso",
			number:      42,
			district:    "Centro",
			propType:    "casa",
			reference:   "REF556",
			yearBuilt:   1950,
			builder:     "Construtora Histórica",
			features:    []string{"jardim histórico", "biblioteca", "varanda colonial", "estacionamento privativo"},
			description: "Residência histórica próxima ao Palácio de Santana, com arquitetura neoclássica preservada e acabamento de luxo",
		},
		{
			name:        "Residencial Porto Digital",
			area:        210.50,
			bedrooms:    3,
			bathrooms:   2,
			garageSpots: 2,
			price:       980000.00,
			street:      "Rua Laranjeiras",
			number:      156,
			district:    "São José",
			propType:    "apartamento",
			reference:   "REF689",
			yearBuilt:   2022,
			builder:     "Inovação Construções",
			features:    []string{"home office", "internet de alta velocidade", "espaço coworking", "sala de reuniões"},
			description: "Apartamento moderno no coração do Porto Digital, ideal para profissionais de tecnologia e startups",
		},
		{
			name:        "Sobrado Rio Sergipe",
			area:        435.20,
			bedrooms:    4,
			bathrooms:   3,
			garageSpots: 3,
			price:       1450000.00,
			street:      "Avenida Rio Sergipe",
			number:      873,
			district:    "Inácio Barbosa",
			propType:    "sobrado",
			reference:   "REF912",
			yearBuilt:   2020,
			builder:     "Sergipe Empreendimentos",
			features:    []string{"deck", "vista para o rio", "jardim tropical", "espaço gourmet"},
			description: "Sobrado espaçoso próximo ao Rio Sergipe, com vista privilegiada e design contemporâneo",
		},
		{
			name:        "Condomínio Universidade",
			area:        195.75,
			bedrooms:    2,
			bathrooms:   2,
			garageSpots: 1,
			price:       620000.00,
			street:      "Avenida Presidente Vargas",
			number:      1105,
			district:    "Centro",
			propType:    "apartamento",
			reference:   "REF045",
			yearBuilt:   2021,
			builder:     "Educacional Construções",
			features:    []string{"segurança 24h", "lavanderia", "espaço de estudos", "wi-fi comum"},
			description: "Apartamento compacto próximo às principais universidades, ideal para estudantes e jovens profissionais",
		},
		{
			name:        "Residência Praia de Atalaia",
			area:        520.30,
			bedrooms:    5,
			bathrooms:   4,
			garageSpots: 4,
			price:       3200000.00,
			street:      "Avenida Presidente Castelo Branco",
			number:      2500,
			district:    "Coroa do Meio",
			propType:    "casa",
			reference:   "REF978",
			yearBuilt:   2022,
			builder:     "Oceânica Incorporações",
			features:    []string{"piscina infinita", "acesso direto à praia", "home theater", "suíte master com varanda"},
			description: "Mansão frente ao mar na Praia de Atalaia, com design arquitetônico único e vistas panorâmicas do oceano",
		},
	}

	props := make([]*properties.Property, 0, len(templates))

	for _, t := range templates {
		photo, blueprint, err := getPhotoAndBlueprint(t.propType)
		if err != nil {
			return nil, fmt.Errorf("could not get photo and blueprint for %s: %w", t.name, err)
		}

		prop := &properties.Property{
			Name:                t.name,
			Area:                t.area,
			NumBedrooms:         t.bedrooms,
			NumBathrooms:        t.bathrooms,
			NumGarageSpots:      t.garageSpots,
			Price:               t.price,
			Street:              t.street,
			Number:              t.number,
			District:            t.district,
			City:                "Aracaju",
			State:               "SE",
			PropertyType:        t.propType,
			Reference:           t.reference,
			PhotoFormat:         strPtr(fmt.Sprintf("%s.jpg", strings.ToLower(strings.ReplaceAll(t.name, " ", "-")))),
			PhotoBase64Data:     strPtr(photo),
			PhotoUploadDate:     &now,
			BlueprintFormat:     strPtr(fmt.Sprintf("%s-blueprint.pdf", strings.ToLower(strings.ReplaceAll(t.name, " ", "-")))),
			BlueprintBase64Data: strPtr(blueprint),
			BlueprintUploadDate: &now,
			Description:         strPtr(t.description),
			YearBuilt:           t.yearBuilt,
			Builder:             strPtr(t.builder),
			Features:            t.features,
			CreatedAt:           now,
			UpdatedAt:           now,
		}
		props = append(props, prop)
	}
	return props, nil
}

func getPhotoAndBlueprint(propType string) (string, string, error) {
	var (
		bpFile  string
		randNum = rand.Intn(2)
	)

	switch randNum {
	case 0:
		bpFile = "assets/planta1.png"
	case 1:
		bpFile = "assets/planta2.png"
	}

	bpRaw, err := assetsFS.ReadFile(bpFile)
	if err != nil {
		return "", "", fmt.Errorf("could not read blueprint file: %w", err)
	}

	var photoFile string

	switch propType {
	case "casa":
		photoFile = "assets/casa1.png"
	case "apartamento":
		randNum = rand.Intn(1)
		switch randNum {
		case 0:
			photoFile = "assets/apartamento1.png"
		case 1:
			photoFile = "assets/apartamento2.png"
		}
	case "sobrado":
		photoFile = "assets/sobrado1.png"
	}

	photoRaw, err := assetsFS.ReadFile(photoFile)
	if err != nil {
		return "", "", fmt.Errorf("could not read photo file: %w", err)
	}
	return base64.StdEncoding.EncodeToString(photoRaw), base64.StdEncoding.EncodeToString(bpRaw), nil
}
