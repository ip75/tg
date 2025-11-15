package domain

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Audio struct {
	MediaID         int // media.id
	Title           string
	Teaser          *string
	Path            string
	MessageThreadID int
	TagID           int // tag.id
	Tag             string
	OccurrenceDate  time.Time
	IssueDate       *time.Time
	Performer       string
	Duration        *time.Duration
	Size            *int
}

func (a Audio) FullLocalPath(basePath string) Audio {
	if len(a.Path) == 0 || filepath.IsAbs(a.Path) {
		return a
	}
	a.Path = filepath.Join(basePath, fmt.Sprintf("%d/%02d", a.OccurrenceDate.Year(), a.OccurrenceDate.Month()), a.Path)
	return a
}

func (a Audio) SetPerformer(p string) Audio {
	a.Performer = p
	return a
}

func (a Audio) Exist() (bool, error) {
	info, err := os.Stat(a.Path)
	if err == nil {
		return !info.IsDir(), nil
	}
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	return false, err
}

func (a Audio) HashTag() string {
	r := strings.NewReplacer("-", "_", " ", "_")
	return fmt.Sprintf("#%s", r.Replace(a.Tag))
}
