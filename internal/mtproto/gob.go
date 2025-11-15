package mtproto

import (
	"bytes"
	"encoding/base64"
	"encoding/gob"
	"fmt"

	"github.com/gotd/td/tg"
)

type audioFile struct {
	tg.InputFile
	Big bool
}

func Marshal(f tg.InputFileClass) (string, error) {

	var (
		res   string
		audio audioFile
	)

	switch f := f.(type) {
	case *tg.InputFile:
		audio = audioFile{
			InputFile: *f,
			Big:       false,
		}
	case *tg.InputFileBig:
		audio = audioFile{
			InputFile: tg.InputFile{
				ID:    f.ID,
				Parts: f.Parts,
				Name:  f.Name,
			},
			Big: true,
		}
	default:
		return "", fmt.Errorf("unknown file type: %T", f)
	}

	var b bytes.Buffer
	if err := gob.NewEncoder(&b).Encode(audio); err != nil {
		return res, fmt.Errorf("serialize: %w", err)
	}

	return base64.RawURLEncoding.EncodeToString(b.Bytes()), nil
}

func Unmarshal(obj string) (tg.InputFileClass, error) {

	var (
		b     bytes.Buffer
		audio audioFile
	)

	d, err := base64.RawURLEncoding.DecodeString(obj)
	if err != nil {
		return nil, fmt.Errorf("base64 decode: %w", err)
	}
	if _, err := b.Write(d); err != nil {
		return nil, fmt.Errorf("put to buffer: %w", err)
	}

	if err := gob.NewDecoder(&b).Decode(&audio); err != nil {
		return nil, fmt.Errorf("deserialize: %w", err)
	}

	if audio.Big {
		return &tg.InputFileBig{
			ID:    audio.ID,
			Parts: audio.Parts,
			Name:  audio.Name,
		}, nil
	}

	return &audio.InputFile, nil
}
