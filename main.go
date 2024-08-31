package main

import (
	"flag"
	"fmt"
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

	ReadFile(file)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
