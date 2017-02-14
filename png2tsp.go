package main

import (
	"math"
	"image"
	//"image/color"
	"github.com/fogleman/gg"
	//"image/png"
	"encoding/binary"
	"flag"
	"fmt"
	"path/filepath"
	"time"
	"io"
	//"io/ioutil"
	"math/rand"
	"os"
	"os/exec"
)

//go:generate stringer -type=FormatType
type FormatType int

const (
	Both FormatType = iota
	PVRTC
	ETC1
)

const PvrtoolPath = "bin\\PVRTexToolCLI.exe"
const EtctoolPath = "bin\\etc1tool.exe"

var tempDir string

var flagFormat = flag.String("f", "etc1", "format (pvr or etc1 or both), default is both")

func check(err error) {
	if err != nil {
		panic(err)
	}
}

var count = 0
func tempPath(file string) string {
	count++
	return filepath.Join(tempDir, fmt.Sprintf("%d%s", count, file))
	//return filepath.Join(tempDir, fmt.Sprintf("%d%s", rand.Int63(), file))
}

func main() {
	flag.Parse()

	var format FormatType
	switch *flagFormat {
	//case "both":
	//	format = Both
	case "pvrtc":
		format = PVRTC
	case "etc1":
		format = ETC1
	default:
		panic(fmt.Sprintf("unknown format %v", *flagFormat))
	}

	if flag.NArg() != 2 {
		flag.PrintDefaults()
		os.Exit(1)
	}
	
	rand.Seed(time.Now().UTC().UnixNano())
	tempDir = filepath.Join(os.TempDir(), "png2tsp")
	os.MkdirAll(tempDir,0666)

	format = ETC1
	args := flag.Args()
	convert(format, args[0], args[1])
}

func upPot(n int) int {
	return int(math.Pow(2, math.Ceil(math.Log2(float64(n)))))
}

func convert(format FormatType, in, out string) {

	img, err := gg.LoadPNG(in)
	check(err)

	bounds := img.Bounds()
	pot_x := upPot(bounds.Size().X)
	pot_y := upPot(bounds.Size().Y)
	fmt.Println(pot_x, pot_y)

	pot_img := gg.NewContext(pot_x, pot_y)
	pot_img.DrawImage(img, 0, 0)
	img = pot_img.Image()

	img = flipY(img)

	tex := image2tex(format, img)
		
	w, err := os.OpenFile(out, os.O_CREATE, 0666)
	check(err)
	
	tex.Write(w)
}

// 上下逆転させる
func flipY(img_ image.Image) image.Image {
	img := img_.(*image.RGBA)
	bounds := img.Bounds()
	w := bounds.Size().X
	h := bounds.Size().Y
	var x,y int
	for y = 0; y<h/2; y++ {
		for x = 0; x<w; x++ {
			col := img.At(x,y)
			img.Set(x,y, img.At(x,h-y))
			img.Set(x,h-y, col)
		}
	}
	return img
}

func image2tex(format FormatType, img image.Image) *PackedTexture {
	tmp_png := tempPath("image2tex.png")
	err := gg.SavePNG(tmp_png, img)
	check(err)
	return png2tex(format, tmp_png)
}

func png2tex(format FormatType, file string) *PackedTexture {
	var tex *PackedTexture
	switch format {
	case PVRTC:
		tex = png2pvrtc(file)
	case ETC1:
		tex = png2etc1(file)
	default:
		panic("invalid format")
	}

	return tex
}

func png2pvrtc(file string) *PackedTexture {
	return nil
}

func png2etc1(file string) *PackedTexture {
	tmp_pkm := tempPath(".pkm")

	output, err := exec.Command(EtctoolPath, file, "-o", tmp_pkm).Output()
	fmt.Println(string(output))
	check(err)

	pkm, err := os.Open(tmp_pkm)
	check(err)
	
	tex := LoadPKM(pkm)
	return tex
}

type TSPHeader struct {
	Magic [4]byte
	FormatType byte
	Width int16
	Height int16
}

type PackedTexture struct {
	Format FormatType
	Width int
	Height int
	Data []byte
}

type PKMHeader struct {
	Magic [4]byte
	Version [2]byte
	Type [2]byte
	Width int16
	Height int16
	OrigWidth int16
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
		Width: int(header.Width),
		Height: int(header.Height),
		Data: buf,
	}
}

func (t *PackedTexture) Write(w io.Writer){
	header := TSPHeader {
		Magic: [4]byte{'T', 'S', 'P', ' '},
		FormatType: byte(t.Format),
		Width: int16(t.Width),
		Height: int16(t.Height),
	}
	err := binary.Write(w, binary.LittleEndian, &header)
	check(err)

	w.Write(t.Data)
}
