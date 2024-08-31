package main

import (
	"errors"
	"fmt"
	"io"
)

var markerMapping = map[int]string{
	0xffd8: "Start of Image",
	0xffe0: "Application Default Header",
	0xffdb: "Quantization Table",
	0xffc0: "Start of Frame",
	0xffc4: "Define Huffman Table",
	0xffda: "Start of Scan",
	0xffd9: "End of Image",
}

func ReadFile(file io.Reader) error {
	buf := make([]byte, 2)

	for {
		_, err := io.ReadFull(file, buf)
		if err != nil {
			if err == io.EOF {
				fmt.Println("End of file reached")
				break
			}
			return err
		}

		marker := int(buf[0])<<8 | int(buf[1])

		markerString, ok := markerMapping[marker]
		if !ok {
			return errors.New("file is not JPEG")
		}
		fmt.Printf("Marker: %s (0x%X)\n", markerString, marker)

		switch marker {
		case 0xffd8:
			continue

		case 0xffd9:
			return nil

		case 0xffda:
			continue

		default:
			lengthBuf := make([]byte, 2)
			_, err := io.ReadFull(file, lengthBuf)
			if err != nil {
				if err == io.EOF {
					fmt.Println("End of file reached")
					break
				}
				return err
			}
			segmentLength := int(lengthBuf[0])<<8 | int(lengthBuf[1])

			_, err = io.CopyN(io.Discard, file, int64(segmentLength-2))
			if err != nil {
				if err == io.EOF {
					fmt.Println("End of file reached")
					break
				}
				return err
			}
		}
	}
	return nil
}
