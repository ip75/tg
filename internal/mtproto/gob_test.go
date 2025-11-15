package mtproto

import (
	"testing"

	"github.com/gotd/td/tg"
	"github.com/stretchr/testify/require"
)

func TestMarshal(t *testing.T) {
	tests := []struct {
		name    string
		input   tg.InputFileClass
		want    string
		wantErr bool
	}{
		{
			name: "Marshal InputFile",
			input: &tg.InputFile{
				ID:    123,
				Parts: 1,
				Name:  "test.txt",
			},
			wantErr: false,
		},
		{
			name: "Marshal InputFileBig",
			input: &tg.InputFileBig{
				ID:    456,
				Parts: 2,
				Name:  "bigfile.txt",
			},
			wantErr: false,
		},
		{
			name:    "Marshal Unknown Type",
			input:   nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Marshal(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.NotEmpty(t, got)
		})
	}
}

func TestUnmarshal(t *testing.T) {
	// Test cases for Unmarshal
	tests := []struct {
		name    string
		input   string
		want    tg.InputFileClass
		wantErr bool
	}{
		{
			name:    "Unmarshal Empty String",
			input:   "",
			wantErr: true,
		},
		{
			name:    "Unmarshal Invalid Base64",
			input:   "invalid-base64",
			wantErr: true,
		},
		{
			name:    "Unmarshal Invalid Gob Data",
			input:   "aGVsbG8=", // "hello" in base64
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Unmarshal(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}

	// Test round-trip for InputFile
	t.Run("Unmarshal InputFile Round-Trip", func(t *testing.T) {
		input := &tg.InputFile{
			ID:    123,
			Parts: 1,
			Name:  "test.txt",
		}
		marshaled, err := Marshal(input)
		require.NoError(t, err)

		unmarshaled, err := Unmarshal(marshaled)
		require.NoError(t, err)

		require.IsType(t, &tg.InputFile{}, unmarshaled)
		require.Equal(t, input.ID, unmarshaled.(*tg.InputFile).ID)
		require.Equal(t, input.Parts, unmarshaled.(*tg.InputFile).Parts)
		require.Equal(t, input.Name, unmarshaled.(*tg.InputFile).Name)
	})

	// Test round-trip for InputFileBig
	t.Run("Unmarshal InputFileBig Round-Trip", func(t *testing.T) {
		input := &tg.InputFileBig{
			ID:    456,
			Parts: 2,
			Name:  "bigfile.txt",
		}
		marshaled, err := Marshal(input)
		require.NoError(t, err)

		unmarshaled, err := Unmarshal(marshaled)
		require.NoError(t, err)

		require.IsType(t, &tg.InputFileBig{}, unmarshaled)
		require.Equal(t, input.ID, unmarshaled.(*tg.InputFileBig).ID)
		require.Equal(t, input.Parts, unmarshaled.(*tg.InputFileBig).Parts)
		require.Equal(t, input.Name, unmarshaled.(*tg.InputFileBig).Name)
	})
}
