package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
)

type KeysOfAdd int

const (
	Up KeysOfAdd = iota
	Down
	Left
	Right
)

type Options int

const (
	Length Options = iota
	Width
	Input
	Output
	MaxIter
	Freq
)

type GetInfo struct {
	Length  uint16
	Width   uint16
	Input   string
	Output  string
	MaxIter int
	Freq    int
	Flag    [6]bool
}

type Pair struct {
	X uint16
	Y uint16
}

func AddPlace(grid *[][]uint64, arguments *GetInfo, key KeysOfAdd) {
	switch key {
	case Up:
		newRow := make([]uint64, arguments.Length)
		*grid = append([][]uint64{newRow}, *grid...)
		arguments.Width++
	case Down:
		newRow := make([]uint64, arguments.Length)
		*grid = append(*grid, newRow)
		arguments.Width++
	case Left:
		for i := uint16(0); i < arguments.Width; i++ {
			(*grid)[i] = append([]uint64{0}, (*grid)[i]...)
		}
		arguments.Length++
	case Right:
		for i := uint16(0); i < arguments.Width; i++ {
			(*grid)[i] = append((*grid)[i], 0)
		}
		arguments.Length++
	}
}

func Parsing(args []string, arguments *GetInfo) {
	for i := 1; i < len(args); i++ {
		if args[i] == "-l" || args[i] == "--length" {
			i++
			length, err := strconv.Atoi(args[i])
			if err != nil || length <= 0 || length > 65535 {
				fmt.Fprintln(os.Stderr, "Invalid length value. Must be between 1 and 65535")
				os.Exit(1)
			}
			arguments.Length = uint16(length)
			arguments.Flag[Length] = true
		} else if args[i] == "-w" || args[i] == "--width" {
			i++
			width, err := strconv.Atoi(args[i])
			if err != nil || width <= 0 || width > 65535 {
				fmt.Fprintln(os.Stderr, "Invalid width value. Must be between 1 and 65535")
				os.Exit(1)
			}
			arguments.Width = uint16(width)
			arguments.Flag[Width] = true
		} else if args[i] == "-i" || args[i] == "--input" {
			i++
			arguments.Input = args[i]
			arguments.Flag[Input] = true
		} else if args[i] == "-o" || args[i] == "--output" {
			i++
			arguments.Output = args[i]
			arguments.Flag[Output] = true
		} else if args[i] == "-m" || args[i] == "--max-iter" {
			i++
			maxIter, err := strconv.Atoi(args[i])
			if err != nil || maxIter < 0 {
				fmt.Fprintln(os.Stderr, "Invalid max-iter value. Must be positive")
				os.Exit(1)
			}
			arguments.MaxIter = maxIter
			arguments.Flag[MaxIter] = true
		} else if args[i] == "-f" || args[i] == "--freq" {
			i++
			freq, err := strconv.Atoi(args[i])
			if err != nil || freq < 0 {
				fmt.Fprintln(os.Stderr, "Invalid freq value. Must be positive")
				os.Exit(1)
			}
			arguments.Freq = freq
			arguments.Flag[Freq] = true
		} else {
			fmt.Fprintln(os.Stderr, "Unknown option:", args[i])
			os.Exit(1)
		}
	}

	for i := 0; i < 4; i++ {
		if !arguments.Flag[i] {
			fmt.Fprintln(os.Stderr, "Required parameters were not entered")
			os.Exit(1)
		}
	}
}

func ReadingInput(inputPath string, grid [][]uint64, arguments GetInfo, countCellsMore4 *int) {
	file, err := os.Open(inputPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Failed to open input file:", inputPath)
		os.Exit(1)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) != 3 {
			continue
		}

		x, err1 := strconv.Atoi(fields[0])
		y, err2 := strconv.Atoi(fields[1])
		value, err3 := strconv.Atoi(fields[2])

		if err1 != nil || err2 != nil || err3 != nil {
			continue
		}

		if x >= 0 && x < int(arguments.Width) && y >= 0 && y < int(arguments.Length) {
			grid[x][y] = uint64(value)
			if value >= 4 {
				*countCellsMore4++
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "Error reading input file:", err)
		os.Exit(1)
	}
}

func Iteration(grid *[][]uint64, arguments *GetInfo, countCellsMore4 *int) {
	ignoring := make(map[Pair]bool)

	for i := uint16(0); i < arguments.Width; i++ {
		for j := uint16(0); j < arguments.Length; j++ {
			temporary := Pair{i, j}
			if _, exists := ignoring[temporary]; exists {
				continue
			}

			if (*grid)[i][j] >= 4 {
				if i == 0 {
					AddPlace(grid, arguments, Up)
					i++
				}
				if j == 0 {
					AddPlace(grid, arguments, Left)
					j++
				}
				if i == arguments.Width-1 {
					AddPlace(grid, arguments, Down)
				}
				if j == arguments.Length-1 {
					AddPlace(grid, arguments, Right)
				}

				(*grid)[i][j] -= 4
				(*grid)[i-1][j]++
				(*grid)[i][j-1]++
				(*grid)[i+1][j]++
				(*grid)[i][j+1]++

				if (*grid)[i][j] < 4 {
					*countCellsMore4--
				}
				if (*grid)[i+1][j] == 4 {
					temporary = Pair{i + 1, j}
					ignoring[temporary] = true
					*countCellsMore4++
				}
				if (*grid)[i][j+1] == 4 {
					temporary = Pair{i, j + 1}
					ignoring[temporary] = true
					*countCellsMore4++
				}
				if (*grid)[i-1][j] == 4 {
					*countCellsMore4++
				}
				if (*grid)[i][j-1] == 4 {
					*countCellsMore4++
				}
			}
		}
	}
}
