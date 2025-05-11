package main

import (
	"os"
)

func main() {
	var arguments GetInfo
	Parsing(os.Args, &arguments)

	grid := make([][]uint64, arguments.Width)
	for i := uint16(0); i < arguments.Width; i++ {
		grid[i] = make([]uint64, arguments.Length)
	}

	countCellsMore4 := 0
	ReadingInput(arguments.Input, grid, arguments, &countCellsMore4)

	nowIterations := 0
	for countCellsMore4 != 0 {
		Iteration(&grid, &arguments, &countCellsMore4)
		nowIterations++

		if arguments.Flag[Freq] {
			if arguments.Freq != 0 && nowIterations%arguments.Freq == 0 && countCellsMore4 != 0 {
				CreateBmp(arguments.Output, &grid, arguments, nowIterations)
			}
		}

		if arguments.Flag[MaxIter] {
			if nowIterations == arguments.MaxIter {
				break
			}
		}
	}

	CreateBmp(arguments.Output, &grid, arguments, nowIterations)
}
