package main

import (
	"fmt"
	"os"
)

func GetPath(path string, iteration int) string {
	filename := fmt.Sprintf("/iter_%04d.bmp", iteration)
	return path + filename
}

func ToNumbers4(x int, numbers4 []byte) {
	numbers4[3] = byte(x / (256 * 256 * 256))
	x = x % (256 * 256 * 256)
	numbers4[2] = byte(x / (256 * 256))
	x = x % (256 * 256)
	numbers4[1] = byte(x / 256)
	numbers4[0] = byte(x % 256)
}

func CreateBmp(path string, basic *[][]uint64, arguments GetInfo, iteration int) {
	filepath := GetPath(path, iteration)
	fout, err := os.Create(filepath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Error creating file:", filepath)
		os.Exit(1)
	}
	defer fout.Close()

	length := arguments.Length
	if length%8 != 0 {
		length = length + 8 - length%8
	}

	numbers4 := make([]byte, 4)
	numbers2 := make([]byte, 2)

	const bitmapFileHeader = 14
	const bitmapInfo = 40
	const tableOfColors = 20
	const biBitCount = 4
	const countOfColors = 5

	numbers2[0] = 0x4D
	numbers2[1] = 0x42
	fout.Write([]byte{numbers2[1], numbers2[0]})

	bfSize := int64(bitmapFileHeader + bitmapInfo + tableOfColors + (int(length)*int(arguments.Width))/2)
	ToNumbers4(int(bfSize), numbers4)
	fout.Write(numbers4)

	numbers2[0] = 0
	numbers2[1] = 0
	fout.Write([]byte{numbers2[0], numbers2[1], numbers2[0], numbers2[1]})

	numbers4[0] = byte(bitmapFileHeader + bitmapInfo + tableOfColors)
	numbers4[1], numbers4[2], numbers4[3] = 0, 0, 0
	fout.Write(numbers4)

	numbers4[0] = byte(bitmapInfo)
	numbers4[1], numbers4[2], numbers4[3] = 0, 0, 0
	fout.Write(numbers4)

	ToNumbers4(int(arguments.Length), numbers4)
	fout.Write(numbers4)

	ToNumbers4(int(arguments.Width), numbers4)
	fout.Write(numbers4)

	numbers2[0] = 1
	numbers2[1] = 0
	fout.Write(numbers2)

	numbers2[0] = byte(biBitCount)
	numbers2[1] = 0
	fout.Write(numbers2)

	numbers4[0], numbers4[1], numbers4[2], numbers4[3] = 0, 0, 0, 0
	fout.Write(numbers4)
	fout.Write(numbers4)

	fout.Write(numbers4)
	fout.Write(numbers4)

	numbers4[0] = byte(countOfColors)
	numbers4[1], numbers4[2], numbers4[3] = 0, 0, 0
	fout.Write(numbers4)
	fout.Write(numbers4)

	colors := [5][4]byte{
		{255, 255, 255, 0},
		{0, 255, 0, 0},
		{128, 0, 128, 0},
		{0, 255, 255, 0},
		{0, 0, 0, 0},
	}

	for i := 0; i < 5; i++ {
		fout.Write([]byte{colors[i][0], colors[i][1], colors[i][2], colors[i][3]})
	}

	for i := int(arguments.Width) - 1; i >= 0; i-- {
		for j := uint16(0); j < length; j += 2 {
			var a, b byte
			if j >= arguments.Length {
				a, b = 0, 0
			} else if j+1 >= arguments.Length {
				if (*basic)[i][j] > 4 {
					a = 4
				} else {
					a = byte((*basic)[i][j])
				}
				b = 0
			} else {
				if (*basic)[i][j] > 4 {
					a = 4
				} else {
					a = byte((*basic)[i][j])
				}

				if (*basic)[i][j+1] > 4 {
					b = 4
				} else {
					b = byte((*basic)[i][j+1])
				}
			}
			q := (a << 4) + b
			fout.Write([]byte{q})
		}
	}
}
