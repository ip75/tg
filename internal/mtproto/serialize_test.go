package mtproto

import (
	"bytes"
	"encoding/gob"
	"testing"

	"github.com/brimdata/super/sup"
	"github.com/gotd/td/constant"
	"github.com/gotd/td/tg"
	"github.com/stretchr/testify/require"
)

func TestTextSerialize(t *testing.T) {

	file := tg.InputFile{
		ID:          1,
		Parts:       123,
		Name:        "test",
		MD5Checksum: "123",
	}

	_ = constant.UploadMaxSmallSize

	fileBig := tg.InputFileBig{
		ID:    2,
		Parts: 321,
		Name:  "big_test",
	}

	var i tg.InputFileClass = &file
	var i2 tg.InputFileClass = &fileBig

	m := sup.NewBSUPMarshaler()
	m.Decorate(sup.StyleFull)

	iRes, err := m.Marshal(i)
	require.NoError(t, err)

	t.Logf("iRes: %s", iRes)

	u := sup.NewBSUPUnmarshaler()
	require.NoError(t, u.Bind(tg.InputFile{}, tg.InputFileBig{}))

	err = u.Unmarshal(iRes, i2)
	require.NoError(t, err)

	t.Logf("i2: %s", i2)
	t.Logf("fileBig: %+v", fileBig)
}

func TestGobSerialize(t *testing.T) {
	t.Skip()
	file := tg.InputFile{
		ID:          1,
		Parts:       123,
		Name:        "test",
		MD5Checksum: "123",
	}

	fileBig := tg.InputFileBig{
		ID:    2,
		Parts: 321,
		Name:  "big_test",
	}

	var i tg.InputFileClass = &file
	var i2 tg.InputFileClass // = &fileBig

	var b bytes.Buffer
	enc := gob.NewEncoder(&b)
	dec := gob.NewDecoder(&b)

	gob.Register(file)
	gob.Register(fileBig)

	err := enc.Encode(i)
	require.NoError(t, err)

	t.Logf("b: %s", b.String())

	err = dec.Decode(i2)
	require.NoError(t, err)

	t.Logf("i: %s", i)
	t.Logf("i2: %s", i2)
	t.Logf("fileBig: %+v", fileBig)
	t.Logf("file: %+v", file)

	type Union struct {
		F   tg.InputFileClass
		Big bool
	}

	u := Union{
		F:   &file,
		Big: false,
	}

	t.Logf("u: %+v", u)

	err = enc.Encode(u)
	require.NoError(t, err)

	t.Logf("b: %s", b.String())

	var u2 Union

	err = dec.Decode(&u2)
	require.NoError(t, err)

	t.Logf("u2: %+v", u2)

}
