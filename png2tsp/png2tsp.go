package main

import (
	"flag"
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"math"
	"math/rand"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/kardianos/osext"
)

//go:generate stringer -type=FormatType
type FormatType int

const (
	Both FormatType = iota
	PVRTC
	ETC1
)

var PvrtoolPath string
var EtctoolPath string

var tempDir string

var optFormat = flag.String("f", "etc1", "Output texture format, PVRTC or ETC1, default is ETC1")
var optDither = flag.Bool("dither", false, "Add dihter to avoid compress artifact, enable only '-f PVRTC'")
var optBatch = flag.Bool("batch", false, "Run as batch mode")
var optPostfix = flag.String("postfix", ".tsp", "Run as batch mode. (Default '.tsp')")
var optOutDir = flag.String("outdir", "", "Output directory")
var optJobs = flag.Int("j", 4, "Parallel job number, enable only batch mode.")

func check(err error) {
	if err != nil {
		panic(err)
	}
}

var count = 0

// メインエントリ
func main() {
	flag.Usage = func() {
		fmt.Println("png2tsp: texture converter for dynamic atlas")
		flag.PrintDefaults()
	}
	flag.Parse()

	var format FormatType
	switch strings.ToUpper(*optFormat) {
	case "":
	case "ETC1":
		format = ETC1
	case "PVRTC":
		format = PVRTC
	default:
		panic(fmt.Sprintf("unknown format %v", *optFormat))
	}

	initialize()

	if *optBatch {
		// バッチモード
		if flag.NArg() <= 0 {
			flag.PrintDefaults()
			os.Exit(1)
		}

		doBatch(format, flag.Args())

	} else {
		// １ファイル変換モード

		if flag.NArg() != 2 {
			flag.PrintDefaults()
			os.Exit(1)
		}

		args := flag.Args()
		convert(format, args[0], args[1])
	}
}

// 初期化を行う
func initialize() {
	// 実行ファイルの場所を取得する
	bindir, err := osext.ExecutableFolder()
	check(err)
	if os.PathSeparator == '\\' {
		PvrtoolPath = filepath.Join(bindir, "PVRTexToolCLI.exe")
		EtctoolPath = filepath.Join(bindir, "etc1tool.exe")
	} else {
		PvrtoolPath = filepath.Join(bindir, "PVRTexToolCLI")
		EtctoolPath = filepath.Join(bindir, "etc1tool")
	}

	// ランダムの初期化
	rand.Seed(time.Now().UTC().UnixNano())

	// Tempディレクトリを作成する
	tempDir = filepath.Join(os.TempDir(), "png2tsp")
	err = os.MkdirAll(tempDir, 0777)
	check(err)
}

// バッチモードで起動する
func doBatch(format FormatType, files []string) {
	wg := sync.WaitGroup{}

	ch := make(chan string)

	for i := 0; i < *optJobs; i++ {
		wg.Add(1)
		go func() {
			for infile := range ch {
				ext := filepath.Ext(infile)
				basename := filepath.Base(infile)
				basename = basename[0 : len(basename)-len(ext)]
				dir := filepath.Dir(infile)
				if *optOutDir != "" {
					dir = *optOutDir
				}
				outfile := filepath.Join(dir, basename+*optPostfix)
				fmt.Printf("converting %v ...\n", outfile)

				convert(format, infile, outfile)
			}
			wg.Done()
		}()
	}

	for _, file := range files {
		ch <- file
	}

	close(ch)

	wg.Wait()
}

// テンポラリファイルのパスを取得する
func tempPath(file string) string {
	if true {
		return filepath.Join(tempDir, fmt.Sprintf("%d%s", rand.Int63(), file))
	} else {
		// デバッグ用の実装
		count++
		return filepath.Join(tempDir, fmt.Sprintf("%d%s", count, file))
	}
}

// 2の塁上に切り上げる
func floorToPowerOf2(n int) int {
	return int(math.Pow(2, math.Ceil(math.Log2(float64(n)))))
}

// ファイルをTSPにコンバートする
func convert(format FormatType, in, out string) {

	inReader, err := os.Open(in)
	check(err)

	img, err := png.Decode(inReader)
	check(err)

	size := img.Bounds().Size()
	pot_x := floorToPowerOf2(size.X)
	pot_y := floorToPowerOf2(size.Y)

	if format == PVRTC {
		pot_x = int(math.Max(float64(pot_x), float64(pot_y)))
		pot_y = pot_x
	}

	pot_img := image.NewRGBA(image.Rect(0, 0, pot_x, pot_y))
	draw.Draw(pot_img, image.Rect(0, pot_y-size.Y, size.X, pot_y), img, image.Point{0, 0}, draw.Src)
	img = pot_img

	img = flipY(img)

	potTex := imageToPackedTexture(format, img)

	blockSizeX := int(math.Ceil(float64(size.X)/4) * 4)
	blockSizeY := int(math.Ceil(float64(size.Y)/4) * 4)

	clipedTex := NewPackedTexture(format, blockSizeX, blockSizeY)
	clipedTex.CopyFrom(potTex, 0, 0, 0, 0, blockSizeX/BlockSize, blockSizeY/BlockSize)

	w, err := os.OpenFile(out, os.O_CREATE|os.O_WRONLY, 0666)
	check(err)

	clipedTex.Write(w)
}

// 画像を上下逆転させる
// Unityが画像を反転させものを使用するため
func flipY(img_ image.Image) image.Image {
	img := img_.(*image.RGBA)
	bounds := img.Bounds()
	w := bounds.Size().X
	h := bounds.Size().Y
	var x, y int
	for y = 0; y < h/2; y++ {
		for x = 0; x < w; x++ {
			col := img.At(x, y)
			img.Set(x, y, img.At(x, h-y))
			img.Set(x, h-y, col)
		}
	}
	return img
}

// Imageをフォーマットを指定して,PackedTextureに変換する.
// その際、TextureをPOTにサイズ変換される。
// PVRTCならWidth,Heightともに同じサイズまで拡張される。
func imageToPackedTexture(format FormatType, img image.Image) *PackedTexture {
	tmp_png := tempPath("image2tex.png")
	writer, err := os.OpenFile(tmp_png, os.O_CREATE|os.O_WRONLY, 0666)
	check(err)

	err = png.Encode(writer, img)
	check(err)

	return pngToPackedTexture(format, tmp_png)
}

// PNGをフォーマットを指定して、PackedTextureに変換する
func pngToPackedTexture(format FormatType, file string) *PackedTexture {
	var tex *PackedTexture
	switch format {
	case PVRTC:
		tex = pngToPVRTC(file)
	case ETC1:
		tex = pngToETC1(file)
	default:
		panic("invalid format")
	}

	return tex
}

// PNGをPVRTCに変換する
func pngToPVRTC(file string) *PackedTexture {
	tmp_pvr := tempPath(".pvr")

	args := []string{"-f", "PVRTC1_4", "-l", "-b8,8", "-i", file, "-o", tmp_pvr}
	if *optDither {
		args = append([]string{"-dither"}, args...)
	}
	output, err := exec.Command(PvrtoolPath, args...).Output()
	if err != nil {
		fmt.Println(string(output))
	}
	check(err)

	pkm, err := os.Open(tmp_pvr)
	check(err)

	tex := LoadPVR(pkm)
	return tex
}

// PNGをETC1に変換する
func pngToETC1(file string) *PackedTexture {
	tmp_pkm := tempPath(".pkm")

	output, err := exec.Command(EtctoolPath, file, "-o", tmp_pkm).Output()
	if err != nil {
		fmt.Println(string(output))
	}
	check(err)

	pkm, err := os.Open(tmp_pkm)
	check(err)

	tex := LoadPKM(pkm)
	return tex
}
