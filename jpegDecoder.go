package main

import (
	"errors"
	"fmt"
	"io"
	"os"
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

func ReadFile(file *os.File) error {
	buf := make([]byte, 2)

	for {
		_, err := file.Read(buf)
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
			_, err := file.Read(lengthBuf)
			if err != nil {
				if err == io.EOF {
					fmt.Println("End of file reached")
					break
				}
				return err
			}

			segmentLength := int(lengthBuf[0])<<8 | int(lengthBuf[1])

			if marker == 0xffc4 {
				decodeHuffman(file)
			} else {

				file.Seek(int64(segmentLength-2), io.SeekCurrent)
			}

		}
	}
	return nil
}

func decodeHuffman(data *os.File) error {
	buf := make([]byte, 2)

	_, err := data.Read(buf)
	if err != nil {
		return err
	}

	len_buf := make([]byte, 16)

	_, err = data.Seek(-1, io.SeekCurrent)
	if err != nil {
		return err
	}

	_, err = data.Read(len_buf)

	if err != nil {
		return err
	}

	elements := []byte{}
	for _, v := range len_buf {
		if v == 0 {
			continue
		}
		elem := make([]byte, v)
		_, err := data.Read(elem)
		if err != nil {
			return err
		}

		elements = append(elements, elem...)
	}

	newHuffman := NewHuffman()

	newHuffman.getHuffmanBits(len_buf, elements)

	fmt.Printf("haeder %d\n", int(buf[0])<<8|int(buf[1]))
	fmt.Printf("length %d\n", len_buf)
	fmt.Printf("elemnts %d\n", len(elements))
	fmt.Printf("roots %v\n", newHuffman.root)
	fmt.Printf("elents %v\n", newHuffman.elemnts)

	return nil
}

type Huffman struct {
	root    []interface{}
	elemnts []byte
}

func NewHuffman() *Huffman {
	return &Huffman{
		root: make([]interface{}, 0),
	}
}

func (hf *Huffman) bitsFromLength(root *[]interface{}, element byte, pos int) bool {
	if len(*root) == 0 {
		*root = append(*root, make([]interface{}, 0))
	}

	if pos == 0 {
		if len(*root) < 2 {
			*root = append(*root, element)
			return true
		}
		return false
	}

	for i := 0; i < 2; i++ {
		if len(*root) == i {
			*root = append(*root, make([]interface{}, 0))
		}
		var childRoot []interface{}
		if existing, ok := (*root)[i].([]interface{}); ok {
			childRoot = existing
		} else {
			childRoot = make([]interface{}, 0)
		}

		if hf.bitsFromLength(&childRoot, element, pos-1) {
			(*root)[i] = childRoot
			return true
		}
	}

	return false
}

func (hf *Huffman) getHuffmanBits(lengths []byte, elements []byte) {
	hf.elemnts = elements

	ii := 0
	for i := 0; i < len(lengths); i++ {
		for _ = range lengths[i] {
			hf.bitsFromLength(&hf.root, elements[ii], i)
			ii++
		}
	}
}
