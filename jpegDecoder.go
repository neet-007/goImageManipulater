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

	huffmanTables := map[int][]byte{}
	QTTables := map[int][]byte{}
	QTMapping := []int{}

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
				decodeHuffman(file, huffmanTables)
				fmt.Printf("huffmens %v\n", huffmanTables)
			} else if marker == 0xffdb {
				defineQuantizationTables(file, QTTables)
				fmt.Printf("qt tables %v\n", QTTables)
			} else if marker == 0xffc0 {
				baselineDCT(file, QTMapping)
				fmt.Printf("components %v\n", QTMapping)
			} else {

				file.Seek(int64(segmentLength-2), io.SeekCurrent)
			}

		}
	}

	return nil
}

func decodeHuffman(data *os.File, huffmanTables map[int][]byte) error {
	buf := make([]byte, 2)

	_, err := data.Read(buf)
	if err != nil {
		return err
	}

	marker := int(buf[0])<<8 | int(buf[1])

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

	huffmanTables[marker] = newHuffman.elemnts

	//fmt.Printf("haeder %d\n", int(buf[0])<<8|int(buf[1]))
	//fmt.Printf("length %d\n", len_buf)
	//fmt.Printf("elemnts %d\n", len(elements))
	//fmt.Printf("roots %v\n", newHuffman.root)
	//fmt.Printf("elents %v\n", newHuffman.elemnts)

	return nil
}

type Huffman struct {
	root    []interface{}
	elemnts []byte
}

type Stream struct {
	data []int
	pos  int
}

func NewHuffman() *Huffman {
	return &Huffman{
		root: make([]interface{}, 0),
	}
}

func NewStream() *Stream {
	return &Stream{
		data: make([]int, 0),
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
			continue
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
		for range lengths[i] {
			hf.bitsFromLength(&hf.root, elements[ii], i)
			ii++
		}
	}
}

func (hf *Huffman) find(st Stream) interface{} {
	var r interface{}
	r = hf.root

	for {
		if slice, ok := r.([]interface{}); ok {
			r = slice[st.getBit()]
		} else {
			break
		}
	}

	return r
}

func (hf *Huffman) getCode(st Stream) interface{} {
	for {
		res := hf.find(st)

		if res != -1 {
			return res
		}
	}
}

func (st *Stream) getBit() int {
	b := st.data[st.pos>>3]
	s := 7 - (st.pos & 0x7)
	st.pos++
	return (b >> s) & 1
}

func (st *Stream) getBitN(l int) int {
	val := 0
	for range l {
		val = val*2 + st.getBit()
	}
	return val
}

func defineQuantizationTables(data *os.File, QTTables map[int][]byte) error {
	buf := make([]byte, 2)

	_, err := data.Read(buf)
	if err != nil {
		return err
	}

	marker := int(buf[0])<<8 | int(buf[1])

	len_buf := make([]byte, 64)

	_, err = data.Seek(-1, io.SeekCurrent)
	if err != nil {
		return err
	}

	_, err = data.Read(len_buf)

	if err != nil {
		return err
	}

	QTTables[marker] = len_buf

	return nil
}

func baselineDCT(data *os.File, QTMapping []int) error {
	buf := make([]byte, 2)

	_, err := data.Read(buf)
	if err != nil {
		return err
	}

	width := make([]byte, 2)

	_, err = data.Seek(-1, io.SeekCurrent)
	if err != nil {
		return err
	}

	_, err = data.Read(width)

	if err != nil {
		return err
	}

	height := make([]byte, 2)

	_, err = data.Seek(-1, io.SeekCurrent)
	if err != nil {
		return err
	}

	_, err = data.Read(height)

	if err != nil {
		return err
	}

	widthInt := int(width[0])<<8 | int(width[1])
	heightInt := int(height[0]) + int(height[1])<<8
	fmt.Printf("img size %v x %v\n", widthInt, heightInt)

	components := make([]byte, 1)

	_, err = data.Seek(-1, io.SeekCurrent)
	if err != nil {
		return err
	}

	_, err = data.Read(components)

	if err != nil {
		return err
	}

	for range components {
		_, err = data.Seek(1, io.SeekCurrent)
		if err != nil {
			return err
		}

		QtbId := make([]byte, 3)

		_, err := data.Read(QtbId)

		if err != nil {
			return err
		}
		result := 0
		for _, byteValue := range QtbId {
			result = (result << 8) | int(byteValue)
		}
		QTMapping = append(QTMapping, result)
	}

	return nil
}
