package issue

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hamba/avro"
)

func TestDecodeEncode(t *testing.T) {
	schema := avro.MustParse(Schema)

	want := Superhero{
		ID:            234765,
		AffiliationID: 9867,
		Name:          "Wolverine",
		Life:          85.25,
		Energy:        32.75,
		Powers: []*Superpower{
			{ID: 2345, Name: "Bone Claws", Damage: 5, Energy: 1.15, Passive: false},
			{ID: 2346, Name: "Regeneration", Damage: -2, Energy: 0.55, Passive: true},
			{ID: 2347, Name: "Adamant skeleton", Damage: -10, Energy: 0, Passive: true},
		},
	}

	superhero := Superhero{}
	err := avro.Unmarshal(schema, Payload, &superhero)
	assert.NoError(t, err)
	assert.Equal(t, want, superhero)

	t.Log(superhero)

	data, err := avro.Marshal(schema, superhero)
	assert.NoError(t, err)
	assert.Equal(t, data, Payload)
}

func TestGenericDecodeEncode(t *testing.T) {
	schema := avro.MustParse(Schema)

	var superhero interface{}
	err := avro.Unmarshal(schema, Payload, &superhero)
	assert.NoError(t, err)

	t.Log(superhero)

	data, err := avro.Marshal(schema, superhero)
	assert.NoError(t, err)
	assert.Equal(t, data, Payload)
}