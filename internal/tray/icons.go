package tray

import (
	"bytes"
	_ "embed"
	"encoding/binary"
	"image"
	"image/png"
)

//go:embed grey.png
var iconGreyPNG []byte

//go:embed green.png
var iconGreenPNG []byte

//go:embed yellow.png
var iconYellowPNG []byte

//go:embed red.png
var iconRedPNG []byte

var (
	IconGrey   []byte
	IconGreen  []byte
	IconYellow []byte
	IconRed    []byte
)

func init() {
	IconGrey = pngToICO(iconGreyPNG)
	IconGreen = pngToICO(iconGreenPNG)
	IconYellow = pngToICO(iconYellowPNG)
	IconRed = pngToICO(iconRedPNG)
}

// pngToICO wraps a PNG in a minimal ICO container.
func pngToICO(pngData []byte) []byte {
	// Decode PNG to get dimensions
	img, err := png.Decode(bytes.NewReader(pngData))
	if err != nil {
		return pngData // fallback to raw PNG
	}

	bounds := img.Bounds()
	w := bounds.Dx()
	h := bounds.Dy()

	// ICO width/height: 0 means 256
	icoW := byte(w)
	icoH := byte(h)
	if w >= 256 {
		icoW = 0
	}
	if h >= 256 {
		icoH = 0
	}

	// Re-encode as PNG to get clean data
	var pngBuf bytes.Buffer
	png.Encode(&pngBuf, img)
	pngBytes := pngBuf.Bytes()

	// Build ICO file
	var ico bytes.Buffer

	// ICONDIR header (6 bytes)
	binary.Write(&ico, binary.LittleEndian, uint16(0))    // reserved
	binary.Write(&ico, binary.LittleEndian, uint16(1))    // type: 1 = ICO
	binary.Write(&ico, binary.LittleEndian, uint16(1))    // count: 1 image

	// ICONDIRENTRY (16 bytes)
	ico.WriteByte(icoW)                                             // width
	ico.WriteByte(icoH)                                             // height
	ico.WriteByte(0)                                                // color palette
	ico.WriteByte(0)                                                // reserved
	binary.Write(&ico, binary.LittleEndian, uint16(1))              // color planes
	binary.Write(&ico, binary.LittleEndian, uint16(32))             // bits per pixel
	binary.Write(&ico, binary.LittleEndian, uint32(len(pngBytes)))  // image size
	binary.Write(&ico, binary.LittleEndian, uint32(22))             // offset (6 + 16)

	// PNG data
	ico.Write(pngBytes)

	return ico.Bytes()
}

// Ensure image package is used
var _ image.Image
