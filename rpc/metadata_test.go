package rpc_test

import (
	"testing"

	"github.com/hamba/avro/rpc"
	"github.com/stretchr/testify/assert"
)

func TestMetadata_Set(t *testing.T) {
	m := rpc.Metadata{}

	m.Set("test", []byte{0x01})

	assert.Equal(t, rpc.Metadata{"test": []byte{0x01}}, m)
}

func TestMetadata_Get(t *testing.T) {
	m := rpc.Metadata{"test": []byte{0x01}}

	got := m.Get("test")

	assert.Equal(t, []byte{0x01}, got)
}

func TestMetadata_GetOnNilInstance(t *testing.T) {
	var m rpc.Metadata

	got := m.Get("test")

	assert.Len(t, got, 0)
}
