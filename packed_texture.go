package main

import (
	"encoding/binary"
	"io"
)

type TSPHeader struct {
	Magic      [4]byte
	FormatType byte
	Width      int16
	Height     int16
}

type PackedTexture struct {
	Format FormatType
	Width  int
	Height int
	Data   []byte
}

type PKMHeader struct {
	Magic      [4]byte
	Version    [2]byte
	Type       [2]byte
	Width      int16
	Height     int16
	OrigWidth  int16
	OrigHeight int16
}

func LoadPKM(r io.Reader) *PackedTexture {
	header := PKMHeader{}

	err := binary.Read(r, binary.BigEndian, &header)
	check(err)

	size := int(header.Width) * int(header.Height) / 2
	buf := make([]byte, size)
	n, err := r.Read(buf)
	check(err)
	if n != len(buf) {
		panic("invalid size")
	}

	// fmt.Printf("%v %v %v %v\n", header.Magic, header.Width, header.Height, len(buf))

	return &PackedTexture{
		Format: ETC1,
		Width:  int(header.Width),
		Height: int(header.Height),
		Data:   buf,
	}
}

func (t *PackedTexture) Write(w io.Writer) {
	header := TSPHeader{
		Magic:      [4]byte{'T', 'S', 'P', ' '},
		FormatType: byte(t.Format),
		Width:      int16(t.Width),
		Height:     int16(t.Height),
	}
	err := binary.Write(w, binary.LittleEndian, &header)
	check(err)

	w.Write(t.Data)
}
