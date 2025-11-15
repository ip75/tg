package domain

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestAudio_FullLocalPath(t *testing.T) {
	tests := []struct {
		name     string
		audio    Audio
		basePath string
		want     Audio
	}{
		{
			name: "Absolute Path",
			audio: Audio{
				Path: "/absolute/path/to/file.mp3",
			},
			basePath: "/base/path",
			want: Audio{
				Path: "/absolute/path/to/file.mp3",
			},
		},
		{
			name: "Relative Path",
			audio: Audio{
				Path:           "relative/path/to/file.mp3",
				OccurrenceDate: time.Date(2025, time.August, 24, 0, 0, 0, 0, time.UTC),
			},
			basePath: "/base/path",
			want: Audio{
				Path:           filepath.Join("/base/path", "2025/08", "relative/path/to/file.mp3"),
				OccurrenceDate: time.Date(2025, time.August, 24, 0, 0, 0, 0, time.UTC),
			},
		},
		{
			name: "Empty Path",
			audio: Audio{
				Path: "",
			},
			basePath: "/base/path",
			want: Audio{
				Path: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.audio.FullLocalPath(tt.basePath)
			require.Equal(t, tt.want.Path, got.Path)
		})
	}
}

func TestAudio_Initialization(t *testing.T) {
	t.Run("Default Values", func(t *testing.T) {
		audio := Audio{}
		require.Equal(t, 0, audio.MediaID)
		require.Equal(t, "", audio.Title)
		require.Equal(t, "", audio.Path)
		require.Equal(t, 0, audio.MessageThreadID)
		require.Equal(t, "", audio.Tag)
		require.Equal(t, time.Time{}, audio.OccurrenceDate)
		require.Nil(t, audio.IssueDate)
	})

	t.Run("With IssueDate", func(t *testing.T) {
		issueDate := time.Now()
		audio := Audio{
			IssueDate: &issueDate,
		}
		require.NotNil(t, audio.IssueDate)
		require.Equal(t, issueDate, *audio.IssueDate)
	})
}

func TestAudio_Exist(t *testing.T) {
	// Create a temporary file for testing
	tmpFile, err := os.CreateTemp("", "testfile")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())

	// Create a temporary directory for testing
	tmpDir, err := os.MkdirTemp("", "testdir")
	require.NoError(t, err)
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name    string
		audio   Audio
		want    bool
		wantErr bool
	}{
		{
			name: "File Exists",
			audio: Audio{
				Path: tmpFile.Name(),
			},
			want:    true,
			wantErr: false,
		},
		{
			name: "File Does Not Exist",
			audio: Audio{
				Path: "nonexistentfile",
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "Directory Path",
			audio: Audio{
				Path: tmpDir,
			},
			want:    false,
			wantErr: false,
		},
		{
			name: "Empty Path",
			audio: Audio{
				Path: "",
			},
			want:    false,
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.audio.Exist()
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
			}
			require.Equal(t, tt.want, got)
		})
	}
}
