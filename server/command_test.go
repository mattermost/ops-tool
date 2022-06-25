package server

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestDedup(t *testing.T) {
	assert := require.New(t)

	res := dedup([]OpsCommandVariable{
		{Name: "a", Value: "va"},
		{Name: "b", Value: "vb"},
		{Name: "a", Value: "va2"},
		{Name: "c", Value: "vc"},
	})

	assert.Len(res, 3)
	assert.Equal("va2", res[0].Value)
	assert.Equal("vb", res[1].Value)
	assert.Equal("vc", res[2].Value)
}
