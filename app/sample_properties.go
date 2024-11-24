package app

import (
	"time"

	"encore.app/properties"
)

func getSampleProperties() []*properties.Property {
	now := time.Now()
	return []*properties.Property{
		{
			ID:             "prop-01",
			Name:           "Mansão Jardins",
			Area:           850.50,
			NumBedrooms:    5,
			NumBathrooms:   6,
			NumGarageSpots: 4,
			Price:          3850000.00,
			Address: properties.Address{
				Street:   "Rua José Steremberg",
				Number:   235,
				District: "Jardins",
				City:     "Aracaju",
				State:    "SE",
			},
			Type: "casa",
			Info: properties.Info{
				Reference: "REF235",
				Photo: properties.ImageData{
					Format:     "jardins-mansion.jpg",
					Base64Data: "", // Placeholder for actual image data
					UploadDate: now,
				},
				Blueprint: properties.ImageData{
					Format:     "jardins-blueprint.pdf",
					Base64Data: "", // Placeholder for actual image data
					UploadDate: now,
				},
				Description: "Mansão luxuosa no bairro mais nobre de Aracaju, com vista privilegiada e acabamento de alto padrão",
				YearBuilt:   2018,
				Builder:     "Construtora Celi",
				Features:    []string{"piscina", "sauna", "quadra de tênis", "churrasqueira"},
			},
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:             "prop-02",
			Name:           "Edifício Le Jardin",
			Area:           245.75,
			NumBedrooms:    4,
			NumBathrooms:   3,
			NumGarageSpots: 3,
			Price:          1250000.00,
			Address: properties.Address{
				Street:   "Avenida Ministro Geraldo Barreto Sobral",
				Number:   1578,
				District: "Grageru",
				City:     "Aracaju",
				State:    "SE",
			},
			Type: "apartamento",
			Info: properties.Info{
				Reference: "REF1578",
				Photo: properties.ImageData{
					Format:     "lejardin-facade.jpg",
					Base64Data: "", // Placeholder for actual image data
					UploadDate: now,
				},
				Blueprint: properties.ImageData{
					Format:     "lejardin-plant.pdf",
					Base64Data: "", // Placeholder for actual image data
					UploadDate: now,
				},
				Description: "Apartamento de alto padrão no Grageru, próximo aos principais shoppings da cidade",
				YearBuilt:   2020,
				Builder:     "Cosil Construções",
				Features:    []string{"piscina", "ginásio", "churrasqueira", "varanda"},
			},
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:             "prop-03",
			Name:           "Residencial Atalaia Sul",
			Area:           320.30,
			NumBedrooms:    3,
			NumBathrooms:   4,
			NumGarageSpots: 2,
			Price:          890000.00,
			Address: properties.Address{
				Street:   "Avenida Santos Dumont",
				Number:   963,
				District: "Atalaia",
				City:     "Aracaju",
				State:    "SE",
			},
			Type: "sobrado",
			Info: properties.Info{
				Reference: "REF-963",
				Photo: properties.ImageData{
					Format:     "atalaia-house.jpg",
					Base64Data: "", // Placeholder for actual image data
					UploadDate: now,
				},
				Blueprint: properties.ImageData{
					Format:     "atalaia-blueprint.pdf",
					Base64Data: "", // Placeholder for actual image data
					UploadDate: now,
				},
				Description: "Sobrado moderno próximo à praia de Atalaia, com vista para o mar",
				YearBuilt:   2019,
				Builder:     "Habitacional Construções",
				Features:    []string{"piscina", "churrasqueira", "varanda"},
			},
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			ID:             "prop-04",
			Name:           "Condomínio Farol da Ilha",
			Area:           178.45,
			NumBedrooms:    3,
			NumBathrooms:   2,
			NumGarageSpots: 2,
			Price:          650000.00,
			Address: properties.Address{
				Street:   "Rua Niceu Dantas",
				Number:   451,
				District: "Farolândia",
				City:     "Aracaju",
				State:    "SE",
			},
			Type: "apartamento",
			Info: properties.Info{
				Reference: "REF451",
				Photo: properties.ImageData{
					Format:     "farol-apto.jpg",
					Base64Data: "", // Placeholder for actual image data
					UploadDate: now,
				},
				Blueprint: properties.ImageData{
					Format:     "farol-plant.pdf",
					Base64Data: "", // Placeholder for actual image data
					UploadDate: now,
				},
				Description: "Apartamento familiar em área universitária, próximo à UNIT e principais supermercados",
				YearBuilt:   2021,
				Builder:     "União Engenharia",
				Features:    []string{"piscina", "ginásio", "varanda"},
			},
			CreatedAt: now,
			UpdatedAt: now,
		},
	}
}
