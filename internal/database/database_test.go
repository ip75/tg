package database

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTgdb_GetSingleInstanceAudio(t *testing.T) {

	d, err := New("postgresql://www@10.108.1.1:5432/bvg")
	require.NoError(t, err)
	defer d.Close()

	got, err := d.GetSingleInstanceAudio(context.Background(), 5)
	require.NoError(t, err)
	assert.NotNil(t, got)

	got, err = d.GetSingleInstanceAudio(context.Background(), 3267)
	require.ErrorIs(t, err, ErrNoSingleInstance)
	assert.Nil(t, got)

	got, err = d.GetSingleInstanceAudio(context.Background(), 100500)
	require.Error(t, err)
	assert.Nil(t, got)

}
