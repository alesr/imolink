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

	return []*properties.Property{
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
