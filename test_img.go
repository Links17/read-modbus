package main

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"github.com/disintegration/imaging"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
)

func main() {
	// 用您的实际base64图像数据替换此占位符
	base64ImageData := "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"

	// 去除MIME前缀并解码base64数据
	//b64data := base64ImageData[strings.IndexByte(base64ImageData, ',')+1:]
	unbased, err := base64.StdEncoding.DecodeString(base64ImageData)
	if err != nil {
		fmt.Println("Error decoding base64:", err)
		os.Exit(1)
	}

	// 解码图像
	imgReader := bytes.NewReader(unbased)
	img, err := imaging.Decode(imgReader)
	if err != nil {
		fmt.Println("Error decoding image with imaging library:", err)
		os.Exit(1)
	}

	// Create a new RGBA image to draw on
	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, image.Point{0, 0}, draw.Src)

	// Define the rectangle's properties
	bboxX := 118 - (97 / 2)
	bboxY := 132 - (104 / 2)
	bboxWidth := 97
	bboxHeight := 104

	// Define the color (red) and line width for the rectangle
	col := color.RGBA{255, 0, 0, 255}
	lineWidth := 2

	// Draw the top line of the rectangle
	for x := bboxX; x < bboxX+bboxWidth; x++ {
		for l := 0; l < lineWidth; l++ {
			rgba.Set(x, bboxY+l, col)
			rgba.Set(x, bboxY+bboxHeight-l, col)
		}
	}

	// Draw the left and right sides of the rectangle
	for y := bboxY; y < bboxY+bboxHeight; y++ {
		for l := 0; l < lineWidth; l++ {
			rgba.Set(bboxX+l, y, col)
			rgba.Set(bboxX+bboxWidth-l, y, col)
		}
	}

	// Encode the image to a new base64 string
	var buff bytes.Buffer
	png.Encode(&buff, rgba)
	newBase64Data := base64.StdEncoding.EncodeToString(buff.Bytes())
	fmt.Println(newBase64Data)
}
