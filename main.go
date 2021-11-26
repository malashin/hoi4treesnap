package main

import (
	"bytes"
	"encoding/gob"
	"image"
	"os"
	"path/filepath"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/widget"
	"github.com/macroblock/imed/pkg/ptool"
	"github.com/malashin/bmfonter"
	_ "github.com/malashin/dds"

	// TGA must be the last registered image format due to not having magic prefix.
	// Every image file will be treated as TGA if registered magic is not found.
	_ "github.com/ftrvxmtrx/tga"
)

var focusTreePaths, modPaths []string
var gamePath, binPath string
var running, isLineRenderingOff bool
var win fyne.Window
var pBar *widget.ProgressBar

var language = "l_english"
var spacingX = 131
var spacingY = 63

var focusMap = make(map[string]Focus)
var gfxMap = make(map[string]SpriteType)
var fontMap = make(map[string]BitmapFont)
var locMap = make(map[string]map[string]Localisation)
var gui FocusGUI
var font, fontTreeTitle bmfonter.Font
var locList, gfxList []string

var buf = new(bytes.Buffer)
var e = gob.NewEncoder(buf)
var d = gob.NewDecoder(buf)

const (
	U Dir = 1
	D Dir = 2
	L Dir = 4
	R Dir = 8
	S Dir = 16
)

var UDdash image.Image
var ULdash image.Image
var URdash image.Image
var DLdash image.Image
var DRdash image.Image
var LRdash image.Image
var UDLdash image.Image
var UDRdash image.Image
var ULRdash image.Image
var DLRdash image.Image
var UDLRdash image.Image
var UD image.Image
var UL image.Image
var UR image.Image
var DL image.Image
var DR image.Image
var LR image.Image
var UDL image.Image
var UDR image.Image
var ULR image.Image
var DLR image.Image
var UDLR image.Image

var pdx *ptool.TParser
var yml *ptool.TParser

var utf8bom = []byte{0xEF, 0xBB, 0xBF}

type Focus struct {
	ID                 string
	Icon               string
	Text               string
	X                  int
	Y                  int
	RelativePositionID string
	Prerequisite       [][]string
	MutuallyExclusive  []string
	AllowBranch        bool
	Available          bool
	Children           []Child
	In                 map[int]FocusLine
	Out                FocusLine
}

type Child struct {
	ID    string
	Solid bool
}

type FocusLine struct {
	Dir Dir
}

type Dir int

type SpriteType struct {
	Name        string
	TextureFile string
	NoOfFrames  int
	Image       image.Image
}

type BitmapFont struct {
	Name      string
	Path      string
	Fontfiles []string
}

type Localisation struct {
	Key    string
	Number string
	Value  string
}

type FocusGUI struct {
	NationalFocusTitle         InstantTextboxType
	NationalFocusItem          ContainerWindowType
	BG                         ButtonType
	Symbol                     ButtonType
	Name                       InstantTextboxType
	NationalFocusLink          ContainerWindowType
	Link                       IconType
	NationalFocusExclusiveItem ContainerWindowType
	Link1                      IconType
	Link2                      IconType
	Left                       IconType
	Right                      IconType
	Mid                        IconType
	FocusSpacing               image.Point
	LinkSpacing                image.Point
	LinkOffsets                image.Point
	LinkBegin                  image.Point
	LinkEnd                    image.Point
	ExclusiveOffset            image.Point
	ExclusiveOffsetLeft        image.Point
	ExclusivePositioning       image.Point
}

type InstantTextboxType struct {
	Name              string
	Position          image.Point
	Orientation       string
	Text              string
	Font              string
	MaxWidth          int
	MaxHeight         int
	Format            string
	VerticalAlignment string
}

type ContainerWindowType struct {
	Name     string
	Position image.Point
	Width    int
	Height   int
}

type ButtonType struct {
	Name           string
	Position       image.Point
	SpriteType     string
	CenterPosition string
	Orientation    string
}

type IconType struct {
	Name       string
	Position   image.Point
	SpriteType string
	Frame      int
}

func main() {
	bin, err := os.Executable()
	if err != nil {
		panic(err)
	}
	binPath = filepath.Dir(bin)

	app := app.New()
	setupUI(app)
}
