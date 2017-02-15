package main

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
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

const BlockSize = 4

func NewPackedTexture(format FormatType, width int, height int) *PackedTexture {
	t := &PackedTexture{
		Format: format,
		Width:  width,
		Height: height,
		Data:   make([]byte, width*height),
	}
	return t
}

func (t *PackedTexture) CopyFrom(src *PackedTexture, src_bx, src_by, dest_bx, dest_by, size_bx, size_by int) {
	src_data := src.Data
	dest_data := t.Data
	for y := 0; y < size_by; y++ {
		for x := 0; x < size_bx; x++ {
			src_idx := ((src_by+y)*src.Width/BlockSize + (src_bx + x)) * 8
			dest_idx := ((dest_by+y)*t.Width/BlockSize + (dest_bx + x)) * 8
			for i := 0; i < 8; i++ {
				dest_data[dest_idx+i] = src_data[src_idx+i]
			}
		}
	}
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

type PVRHeader struct {
	Magic        [4]byte
	Flags        int32
	PixelFormat  [2]int32
	ColorSpec    int32
	ChannelType  int32
	Height       int32
	Width        int32
	Depth        int32
	Surface      int32
	Faces        int32
	MipmapCount  int32
	Metadatasize int32
}

func LoadPVR(r io.Reader) *PackedTexture {
	header := PVRHeader{}

	err := binary.Read(r, binary.LittleEndian, &header)
	check(err)

	if string(header.Magic[:]) != "PVR\x03" {
		panic("invalid PVR header")
	}

	if header.Flags != 0 ||
		header.PixelFormat[0] != 3 ||
		header.PixelFormat[1] != 0 ||
		header.ColorSpec != 0 ||
		header.ChannelType != 0 ||
		header.Depth != 1 ||
		header.Surface != 1 ||
		header.Faces != 1 ||
		header.MipmapCount != 1 {
		fmt.Printf("header %v\n", header)
		panic("invalid PVR header")
	}

	metadata := make([]byte, header.Metadatasize)
	n, err := r.Read(metadata)
	check(err)
	if n != len(metadata) {
		panic("invalid metadata size")
	}

	size := int(header.Width) * int(header.Height) / 2
	buf := make([]byte, size)
	n, err = r.Read(buf)
	check(err)
	if n != len(buf) {
		panic("invalid size")
	}

	// fmt.Printf("%v %v %v %v\n", header, header.Magic, header.Width, header.Height, len(buf))

	// PVRのフォーマットに基づく、配列の入れ替えを行う
	buf2 := make([]byte, size)
	block_width := int(header.Width) / 4
	block_height := int(header.Height) / 4
	bitsize := int(math.Log2(float64(header.Width / 4)))
	for x := 0; x < block_width; x++ {
		for y := 0; y < block_height; y++ {
			src_idx := pvrBlockIdx(bitsize, x, y) * 8
			dest_idx := (y*block_width + x) * 8
			for i := 0; i < 8; i++ {
				buf2[dest_idx+i] = buf[src_idx+i]
			}
		}
	}

	return &PackedTexture{
		Format: PVRTC,
		Width:  int(header.Width),
		Height: int(header.Height),
		Data:   buf2,
	}
}

func pvrBlockIdx(size, bx_, by_ int) int {
	bx := uint(bx_)
	by := uint(by_)
	r := uint(0)
	for i := 0; i < size; i++ {
		r |= (((bx & 1) << 1) | (by & 1)) << uint(i*2)
		bx = bx >> 1
		by = by >> 1
	}
	return int(r)
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
