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
	PVRTC_SPLIT_ALPHA
	ETC1_SPLIT_ALPHA
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
	case "ETC1_SPLIT_ALPHA":
		format = ETC1_SPLIT_ALPHA
	case "PVRTC_SPLIT_ALPHA":
		format = PVRTC_SPLIT_ALPHA
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
			defer func() {
				wg.Done()
			}()
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
	if false {
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

func (f FormatType) SplitAlphaInfo() (bool, FormatType) {
	switch f {
	case PVRTC_SPLIT_ALPHA:
		return true, PVRTC
	case ETC1_SPLIT_ALPHA:
		return true, ETC1
	default:
		return false, f
	}
	panic("must not reach")
}

// ファイルをTSPにコンバートする
func convert(format FormatType, in, out string) {

	r, err := os.Open(in)
	check(err)

	img0, err := png.Decode(r)
	check(err)

	// image.NRGBAに変換する
	img := image.NewNRGBA(img0.Bounds())
	draw.Draw(img, img.Bounds(), img0, image.Point{0, 0}, draw.Src)

	var tex *PackedTexture = nil
	isSplitAlpha, origFormat := format.SplitAlphaInfo()
	if isSplitAlpha {
		// アルファ情報を分ける場合
		// 縦が２倍のテクスチャにアルファと色情報を分けて保存される

		alphaImg := splitAlpha(img)
		alphaTex := imageToPackedTexture(origFormat, alphaImg)

		colorImg := splitColor(img)
		colorTex := imageToPackedTexture(origFormat, colorImg)

		// alphaとcolorに分けたテクスチャを作成する
		tex = NewPackedTexture(format, alphaTex.Width, alphaTex.Height*2)
		blockSizeX := alphaTex.Width / 4
		blockSizeY := alphaTex.Height / 4
		tex.CopyFrom(colorTex, 0, 0, 0, 0, blockSizeX, blockSizeY)
		tex.CopyFrom(alphaTex, 0, 0, 0, blockSizeY, blockSizeX, blockSizeY)

	} else {
		// アルファ情報を分ない場合
		tex = imageToPackedTexture(format, img)
	}

	w, err := os.OpenFile(out, os.O_CREATE|os.O_WRONLY, 0666)
	check(err)

	tex.Write(w)
}

// imageを二の累乗サイズに変更する
// PVRTCなら、縦横を合わせる
func imageToPOT(format FormatType, img *image.NRGBA) *image.NRGBA {
	// POTサイズを取得する
	size := img.Bounds().Size()
	potX := floorToPowerOf2(size.X)
	potY := floorToPowerOf2(size.Y)

	// PVRTCなら、正方形にする
	if format == PVRTC {
		potX = int(math.Max(float64(potX), float64(potY)))
		potY = potX
	}

	// ImageをPOTに変換する
	potImg := image.NewNRGBA(image.Rect(0, 0, potX, potY))
	draw.Draw(potImg, image.Rect(0, potY-size.Y, size.X, potY), img, image.Point{0, 0}, draw.Src)
	return potImg
}

// 画像を上下逆転させる
// Unityが画像を反転させものを使用するため
func flipY(img *image.NRGBA) *image.NRGBA {
	size := img.Bounds().Size()
	w := size.X
	h := size.Y
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

// アルファ成分のみの画像を作成する
// RGBA <= AAA0
func splitAlpha(img *image.NRGBA) *image.NRGBA {
	size := img.Bounds().Size()
	alphaImg := image.NewNRGBA(image.Rect(0, 0, size.X, size.Y))
	length := size.X * size.Y * 4
	for i := 0; i < length; i += 4 {
		alphaImg.Pix[i+0] = img.Pix[i+3]
		alphaImg.Pix[i+1] = img.Pix[i+3]
		alphaImg.Pix[i+2] = img.Pix[i+3]
		alphaImg.Pix[i+3] = 255
	}
	return alphaImg
}

// カラー成分のみの画像を作成する
// RGBA <= RGB1
func splitColor(img *image.NRGBA) *image.NRGBA {
	size := img.Bounds().Size()
	alphaImg := image.NewNRGBA(image.Rect(0, 0, size.X, size.Y))
	length := size.X * size.Y * 4
	for i := 0; i < length; i += 4 {
		alphaImg.Pix[i+0] = img.Pix[i+0]
		alphaImg.Pix[i+1] = img.Pix[i+1]
		alphaImg.Pix[i+2] = img.Pix[i+2]
		alphaImg.Pix[i+3] = 255
	}
	return alphaImg
}

// Imageをフォーマットを指定して,PackedTextureに変換する.
func imageToPackedTexture(format FormatType, img *image.NRGBA) *PackedTexture {
	origSize := img.Bounds().Size() // 元画像のサイズ

	img = imageToPOT(format, img) // POTにする

	img = flipY(img) // Unity向けに上下逆転させる

	//pngに書き出す
	tmp_png := tempPath("image2tex.png")
	writer, err := os.OpenFile(tmp_png, os.O_CREATE|os.O_WRONLY, 0666)
	check(err)

	err = png.Encode(writer, img)
	check(err)

	potTex := pngToPackedTexture(format, tmp_png)

	// 元のサイズで読み込む
	blockSizeX := int(math.Ceil(float64(origSize.X)/4) * 4)
	blockSizeY := int(math.Ceil(float64(origSize.Y)/4) * 4)

	clipedTex := NewPackedTexture(format, blockSizeX, blockSizeY)
	clipedTex.CopyFrom(potTex, 0, 0, 0, 0, blockSizeX/BlockSize, blockSizeY/BlockSize)

	return clipedTex
}

// PNGをフォーマットを指定して、PackedTextureに変換する
func pngToPackedTexture(format FormatType, file string) *PackedTexture {
	var tex *PackedTexture
	switch format {
	case PVRTC:
		tex = pngToPVRTC(format, file)
	case ETC1:
		tex = pngToETC1(file)
	default:
		panic("invalid format")
	}

	return tex
}

// PNGをPVRTCに変換する
func pngToPVRTC(format FormatType, file string) *PackedTexture {
	tmp_pvr := tempPath(".pvr")

	var formatStr string
	if format == PVRTC {
		formatStr = "PVRTC1_4"
	} else {
		formatStr = "PVRTC1_4_RGB"
	}
	args := []string{"-f", formatStr, "-l", "-b8,8", "-i", file, "-o", tmp_pvr}
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
