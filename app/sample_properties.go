package app

import (
	"encoding/base64"
	"fmt"
	"time"

	"math/rand"

	"encore.app/properties"
)

func getSampleProperties() ([]*properties.Property, error) {
	now := time.Now()

	// Reuse photo and blueprint generation logic from previous implementations
	casaPhoto, casaBp, err := getPhotoAndBlueprint("casa")
	if err != nil {
		return nil, fmt.Errorf("could not get photo and blueprint for casa: %w", err)
	}

	apt1Photo, apt1Bp, err := getPhotoAndBlueprint("apartamento")
	if err != nil {
		return nil, fmt.Errorf("could not get photo and blueprint for apartamento 1: %w", err)
	}

	apt2Photo, apt2Bp, err := getPhotoAndBlueprint("apartamento")
	if err != nil {
		return nil, fmt.Errorf("could not get photo and blueprint for apartamento 2: %w", err)
	}

	sobradoPhoto, sobradoBp, err := getPhotoAndBlueprint("sobrado")
	if err != nil {
		return nil, fmt.Errorf("could not get photo and blueprint for sobrado: %w", err)
	}

	// Additional property photos and blueprints
	prop1Photo, prop1Bp, err := getPhotoAndBlueprint("casa")
	if err != nil {
		return nil, fmt.Errorf("could not get photo and blueprint for additional property 1: %w", err)
	}

	prop2Photo, prop2Bp, err := getPhotoAndBlueprint("apartamento")
	if err != nil {
		return nil, fmt.Errorf("could not get photo and blueprint for additional property 2: %w", err)
	}

	prop3Photo, prop3Bp, err := getPhotoAndBlueprint("sobrado")
	if err != nil {
		return nil, fmt.Errorf("could not get photo and blueprint for additional property 3: %w", err)
	}

	prop4Photo, prop4Bp, err := getPhotoAndBlueprint("apartamento")
	if err != nil {
		return nil, fmt.Errorf("could not get photo and blueprint for additional property 4: %w", err)
	}

	prop5Photo, prop5Bp, err := getPhotoAndBlueprint("casa")
	if err != nil {
		return nil, fmt.Errorf("could not get photo and blueprint for additional property 5: %w", err)
	}

	return []*properties.Property{
		// Original 4 properties from the first implementation
		{
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
				Reference: "REF123",
				Photo: properties.ImageData{
					Format:     "jardins-mansion.jpg",
					Base64Data: casaPhoto,
					UploadDate: now,
				},
				Blueprint: properties.ImageData{
					Format:     "jardins-blueprint.pdf",
					Base64Data: casaBp,
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
				Reference: "REF345",
				Photo: properties.ImageData{
					Format:     "lejardin-facade.jpg",
					Base64Data: apt1Photo,
					UploadDate: now,
				},
				Blueprint: properties.ImageData{
					Format:     "lejardin-plant.pdf",
					Base64Data: apt1Bp,
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
				Reference: "REF678",
				Photo: properties.ImageData{
					Format:     "atalaia-house.jpg",
					Base64Data: sobradoPhoto,
					UploadDate: now,
				},
				Blueprint: properties.ImageData{
					Format:     "atalaia-blueprint.pdf",
					Base64Data: sobradoBp,
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
				Reference: "REF901",
				Photo: properties.ImageData{
					Format:     "farol-apto.jpg",
					Base64Data: apt2Photo,
					UploadDate: now,
				},
				Blueprint: properties.ImageData{
					Format:     "farol-plant.pdf",
					Base64Data: apt2Bp,
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
		{
			Name:           "Palácio de Santana",
			Area:           620.80,
			NumBedrooms:    4,
			NumBathrooms:   5,
			NumGarageSpots: 3,
			Price:          2750000.00,
			Address: properties.Address{
				Street:   "Praça Fausto Cardoso",
				Number:   42,
				District: "Centro",
				City:     "Aracaju",
				State:    "SE",
			},
			Type: "casa",
			Info: properties.Info{
				Reference: "REF456",
				Photo: properties.ImageData{
					Format:     "palacio-santana.jpg",
					Base64Data: prop1Photo,
					UploadDate: now,
				},
				Blueprint: properties.ImageData{
					Format:     "palacio-blueprint.pdf",
					Base64Data: prop1Bp,
					UploadDate: now,
				},
				Description: "Residência histórica próxima ao Palácio de Santana, com arquitetura neoclássica preservada e acabamento de luxo",
				YearBuilt:   1950,
				Builder:     "Construtora Histórica",
				Features:    []string{"jardim histórico", "biblioteca", "varanda colonial", "estacionamento privativo"},
			},
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			Name:           "Residencial Porto Digital",
			Area:           210.50,
			NumBedrooms:    3,
			NumBathrooms:   2,
			NumGarageSpots: 2,
			Price:          980000.00,
			Address: properties.Address{
				Street:   "Rua Laranjeiras",
				Number:   156,
				District: "São José",
				City:     "Aracaju",
				State:    "SE",
			},
			Type: "apartamento",
			Info: properties.Info{
				Reference: "REF789",
				Photo: properties.ImageData{
					Format:     "porto-digital-apto.jpg",
					Base64Data: prop2Photo,
					UploadDate: now,
				},
				Blueprint: properties.ImageData{
					Format:     "porto-digital-plant.pdf",
					Base64Data: prop2Bp,
					UploadDate: now,
				},
				Description: "Apartamento moderno no coração do Porto Digital, ideal para profissionais de tecnologia e startups",
				YearBuilt:   2022,
				Builder:     "Inovação Construções",
				Features:    []string{"home office", "internet de alta velocidade", "espaço coworking", "sala de reuniões"},
			},
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			Name:           "Sobrado Rio Sergipe",
			Area:           435.20,
			NumBedrooms:    4,
			NumBathrooms:   3,
			NumGarageSpots: 3,
			Price:          1450000.00,
			Address: properties.Address{
				Street:   "Avenida Rio Sergipe",
				Number:   873,
				District: "Inácio Barbosa",
				City:     "Aracaju",
				State:    "SE",
			},
			Type: "sobrado",
			Info: properties.Info{
				Reference: "REF012",
				Photo: properties.ImageData{
					Format:     "rio-sergipe-sobrado.jpg",
					Base64Data: prop3Photo,
					UploadDate: now,
				},
				Blueprint: properties.ImageData{
					Format:     "rio-sergipe-blueprint.pdf",
					Base64Data: prop3Bp,
					UploadDate: now,
				},
				Description: "Sobrado espaçoso próximo ao Rio Sergipe, com vista privilegiada e design contemporâneo",
				YearBuilt:   2020,
				Builder:     "Sergipe Empreendimentos",
				Features:    []string{"deck", "vista para o rio", "jardim tropical", "espaço gourmet"},
			},
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			Name:           "Condomínio Universidade",
			Area:           195.75,
			NumBedrooms:    2,
			NumBathrooms:   2,
			NumGarageSpots: 1,
			Price:          620000.00,
			Address: properties.Address{
				Street:   "Avenida Presidente Vargas",
				Number:   1105,
				District: "Centro",
				City:     "Aracaju",
				State:    "SE",
			},
			Type: "apartamento",
			Info: properties.Info{
				Reference: "REF345",
				Photo: properties.ImageData{
					Format:     "universidade-apto.jpg",
					Base64Data: prop4Photo,
					UploadDate: now,
				},
				Blueprint: properties.ImageData{
					Format:     "universidade-plant.pdf",
					Base64Data: prop4Bp,
					UploadDate: now,
				},
				Description: "Apartamento compacto próximo às principais universidades, ideal para estudantes e jovens profissionais",
				YearBuilt:   2021,
				Builder:     "Educacional Construções",
				Features:    []string{"segurança 24h", "lavanderia", "espaço de estudos", "wi-fi comum"},
			},
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			Name:           "Residência Praia de Atalaia",
			Area:           520.30,
			NumBedrooms:    5,
			NumBathrooms:   4,
			NumGarageSpots: 4,
			Price:          3200000.00,
			Address: properties.Address{
				Street:   "Avenida Presidente Castelo Branco",
				Number:   2500,
				District: "Coroa do Meio",
				City:     "Aracaju",
				State:    "SE",
			},
			Type: "casa",
			Info: properties.Info{
				Reference: "REF678",
				Photo: properties.ImageData{
					Format:     "atalaia-residencia.jpg",
					Base64Data: prop5Photo,
					UploadDate: now,
				},
				Blueprint: properties.ImageData{
					Format:     "atalaia-residencia-blueprint.pdf",
					Base64Data: prop5Bp,
					UploadDate: now,
				},
				Description: "Mansão frente ao mar na Praia de Atalaia, com design arquitetônico único e vistas panorâmicas do oceano",
				YearBuilt:   2022,
				Builder:     "Oceânica Incorporações",
				Features:    []string{"piscina infinita", "acesso direto à praia", "home theater", "suíte master com varanda"},
			},
			CreatedAt: now,
			UpdatedAt: now,
		},
	}, nil
}

func getPhotoAndBlueprint(propType string) (string, string, error) {
	var (
		bpFile  string
		randNum = rand.Intn(2)
	)

	fmt.Println("randNum1)", randNum)
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
		fmt.Println("randNum)", randNum)
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
