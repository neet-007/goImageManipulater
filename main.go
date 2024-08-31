package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"math"
	"os"
	"strings"
)

func main() {

	flag.Parse()

	args := flag.Args()

	if len(args) <= 0 {
		fmt.Println("must provide image location")
		return
	}

	imageLocation := args[0]
	imageLocationSplit := strings.Split(imageLocation, "/")
	imageName := imageLocationSplit[len(imageLocationSplit)-1]
	imageNameSplit := strings.Split(imageName, ".")

	if len(imageNameSplit) != 2 {
		fmt.Println("image name must be in format name.format")
		return
	}

	if imageNameSplit[1] != "jpeg" && imageNameSplit[1] != "jpg" {
		fmt.Println("image must be in jpg foramt")
		return
	}

	file, err := os.Open(imageLocation)
	check(err)
	defer file.Close()

	img, err := jpeg.Decode(file)
	check(err)

	const TARGET_WIDTH = 1378

	bounds := img.Bounds()

	resizedWidth := bounds.Max.X / 8
	resizedHeight := bounds.Max.Y / 8

	rgba := image.NewRGBA(image.Rect(0, 0, resizedWidth, resizedHeight))

	ASCI_CHARS := [10]string{" ", ".", ";", "c", "o", "P", "O", "?", "%", "#"}
	asciArr := [][]string{}

	for y := 0; y < resizedHeight; y++ {
		asciArr = append(asciArr, []string{})
		for x := 0; x < resizedWidth; x++ {
			asciArr[y] = append(asciArr[y], " ")
		}
	}

	for y := 0; y < resizedHeight; y++ {
		for x := 0; x < resizedWidth; x++ {
			srcX := x * bounds.Dx() / resizedWidth
			srcY := y * bounds.Dy() / resizedHeight

			rgba.Set(x, y, img.At(srcX, srcY))
		}
	}

	for y := 0; y < resizedHeight; y++ {
		for x := 0; x < resizedWidth; x++ {
			originalColor := rgba.At(x, y).(color.RGBA)
			grayScaleColor := math.Floor((0.299*float64(originalColor.R))+
				(0.587*float64(originalColor.G))+
				(0.114*float64(originalColor.B))) / 255 * float64(len(ASCI_CHARS)-1)

			charIndex := int(grayScaleColor)
			if charIndex >= len(ASCI_CHARS) {
				charIndex = len(ASCI_CHARS) - 1
			}
			asciArr[y][x] = ASCI_CHARS[charIndex]

			rgba.Set(x, y, color.RGBA{
				R: uint8(grayScaleColor),
				G: uint8(grayScaleColor),
				B: uint8(grayScaleColor),
			})
		}
	}

	for y := 0; y < resizedHeight; y++ {
		for x := 0; x < resizedWidth; x++ {
			fmt.Print(asciArr[y][x])
		}
		fmt.Println()
	}

	outputFile, err := os.Create("output.jpg")
	if err != nil {
		fmt.Println("Error creating output file:", err)
		return
	}
	defer outputFile.Close()

	err = jpeg.Encode(outputFile, rgba, nil)
	if err != nil {
		fmt.Println("Error encoding JPEG:", err)
		return
	}

	fmt.Println("Image manipulation complete. Saved as output.jpg")

}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
