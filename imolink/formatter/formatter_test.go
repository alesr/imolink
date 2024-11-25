package formatter

import (
	"encoding/json"
	"testing"

	"encore.app/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFormatProperties(t *testing.T) {
	builder := "Test Builder"
	desc := "Test Description"

	props := []*domain.Property{
		{
			Reference:      "REF123",
			PropertyType:   "Apartment",
			Name:           "Test Property",
			Price:          250000.00,
			Street:         "Test Street",
			Number:         123,
			District:       "Test District",
			City:           "Test City",
			State:          "Test State",
			Area:           100.5,
			NumBedrooms:    3,
			NumBathrooms:   2,
			NumGarageSpots: 1,
			YearBuilt:      2020,
			Builder:        &builder,
			Features:       []string{"Pool", "Garden"},
			Description:    &desc,
		},
	}

	result := FormatProperties(props)

	var parsed map[string]any
	err := json.Unmarshal([]byte(result), &parsed)
	require.NoError(t, err)

	properties, ok := parsed["properties"].([]any)
	assert.True(t, ok)
	assert.Len(t, properties, 1)

	// Verify the first property
	prop := properties[0].(map[string]any)
	assert.Equal(t, "REF123", prop["referencia"])
	assert.Equal(t, "Apartment", prop["tipo_imovel"])
	assert.Equal(t, 250000.00, prop["preco"])

	// Verify nested location object
	location := prop["localizacao"].(map[string]any)
	assert.Equal(t, "Test Street", location["rua"])
	assert.Equal(t, float64(123), location["numero"])
}
