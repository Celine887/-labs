package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

type Role int

const (
	NONE Role = iota
	MASTER
	SLAVE
)

type ShotResult int

const (
	MISS ShotResult = iota
	HIT
	KILL
)

type Ship struct {
	Size       int
	IsVertical bool
	X, Y       int
	Hits       int
}

type Pair struct {
	X, Y int
}

type Game struct {
	role           Role
	width, height  int
	shipCounts     map[int]int
	field          [][]rune
	ships          []Ship
	shotHistory    []Pair
	strategy       string
	gameStarted    bool
	allShipsPlaced bool
	isGameCreated  bool
	nextShotX      int
	nextShotY      int
	lastShotResult ShotResult
	hits           []Pair
}

func NewGame() *Game {
	rand.Seed(time.Now().UnixNano())
	return &Game{
		role:           NONE,
		width:          0,
		height:         0,
		shipCounts:     make(map[int]int),
		field:          nil,
		ships:          make([]Ship, 0),
		shotHistory:    make([]Pair, 0),
		strategy:       "custom",
		gameStarted:    false,
		allShipsPlaced: false,
		isGameCreated:  false,
		nextShotX:      0,
		nextShotY:      0,
		lastShotResult: MISS,
		hits:           make([]Pair, 0),
	}
}

func (g *Game) IsValidCoordinate(x, y int) bool {
	return x >= 0 && x < g.width && y >= 0 && y < g.height
}

func (g *Game) CanPlaceShip(size, x, y int, isVertical bool) bool {
	if isVertical {
		if y+size > g.height {
			return false
		}
		for i := 0; i < size; i++ {
			if g.field[y+i][x] != '.' {
				return false
			}
		}
	} else {
		if x+size > g.width {
			return false
		}
		for i := 0; i < size; i++ {
			if g.field[y][x+i] != '.' {
				return false
			}
		}
	}
	return true
}

func (g *Game) PlaceShip(size, x, y int, isVertical bool) {
	newShip := Ship{
		Size:       size,
		IsVertical: isVertical,
		X:          x,
		Y:          y,
		Hits:       0,
	}
	g.ships = append(g.ships, newShip)
	if isVertical {
		for i := 0; i < size; i++ {
			g.field[y+i][x] = 'S'
		}
	} else {
		for i := 0; i < size; i++ {
			g.field[y][x+i] = 'S'
		}
	}
}

func (g *Game) RandomizeShipPlacement() error {
	g.ships = make([]Ship, 0)
	g.field = make([][]rune, g.height)
	for i := range g.field {
		g.field[i] = make([]rune, g.width)
		for j := range g.field[i] {
			g.field[i][j] = '.'
		}
	}

	totalShipCells := 0
	for size, count := range g.shipCounts {
		totalShipCells += size * count
	}
	if totalShipCells > g.width*g.height {
		return fmt.Errorf("Error: not enough space for all ships")
	}

	for size := 1; size <= 4; size++ {
		count := g.shipCounts[size]
		for i := 0; i < count; i++ {
			placed := false
			attempts := 0
			for !placed && attempts < 100 {
				x := rand.Intn(g.width)
				y := rand.Intn(g.height)
				isVertical := rand.Intn(2) == 0
				if g.IsValidCoordinate(x, y) && g.CanPlaceShip(size, x, y, isVertical) {
					g.PlaceShip(size, x, y, isVertical)
					placed = true
				}
				attempts++
			}
			if !placed {
				return fmt.Errorf("Failed to place ships")
			}
		}
	}
	return nil
}

func (g *Game) HandleStartCommand() string {
	if g.width <= 0 || g.height <= 0 {
		return "failed"
	}

	hasShips := false
	for i := 1; i <= 4; i++ {
		if g.shipCounts[i] > 0 {
			hasShips = true
			break
		}
	}
	if !hasShips {
		return "failed"
	}

	totalShipsPlaced := len(g.ships)
	totalShipsCount := 0
	for i := 1; i <= 4; i++ {
		totalShipsCount += g.shipCounts[i]
	}
	if totalShipsPlaced < totalShipsCount {
		if !g.allShipsPlaced {
			err := g.RandomizeShipPlacement()
			if err != nil {
				return "failed"
			}
			g.allShipsPlaced = true
		}
	}
	if !g.isGameCreated {
		return "failed"
	}
	g.gameStarted = true
	return "ok"
}

func (g *Game) SaveToFile(path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	_, err = fmt.Fprintf(writer, "%d %d\n", g.width, g.height)
	if err != nil {
		return err
	}

	for _, ship := range g.ships {
		orientation := "h"
		if ship.IsVertical {
			orientation = "v"
		}
		_, err = fmt.Fprintf(writer, "%d %s %d %d\n", ship.Size, orientation, ship.X, ship.Y)
		if err != nil {
			return err
		}
	}

	return writer.Flush()
}

func (g *Game) LoadFromFile(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	if !scanner.Scan() {
		return fmt.Errorf("Failed to read field dimensions")
	}

	fields := strings.Fields(scanner.Text())
	if len(fields) != 2 {
		return fmt.Errorf("Invalid field dimensions format")
	}

	width, err := strconv.Atoi(fields[0])
	if err != nil {
		return fmt.Errorf("Invalid width: %v", err)
	}

	height, err := strconv.Atoi(fields[1])
	if err != nil {
		return fmt.Errorf("Invalid height: %v", err)
	}

	if width <= 0 || height <= 0 {
		return fmt.Errorf("Invalid field dimensions: %dx%d", width, height)
	}

	g.width = width
	g.height = height
	g.ships = make([]Ship, 0)
	g.shipCounts = make(map[int]int)

	g.field = make([][]rune, g.height)
	for i := range g.field {
		g.field[i] = make([]rune, g.width)
		for j := range g.field[i] {
			g.field[i][j] = '.'
		}
	}

	for scanner.Scan() {
		line := scanner.Text()
		fields := strings.Fields(line)
		if len(fields) != 4 {
			return fmt.Errorf("Invalid ship format: %s", line)
		}

		size, err := strconv.Atoi(fields[0])
		if err != nil {
			return fmt.Errorf("Invalid ship size: %v", err)
		}

		orientation := fields[1]
		isVertical := orientation == "v"

		x, err := strconv.Atoi(fields[2])
		if err != nil {
			return fmt.Errorf("Invalid X coordinate: %v", err)
		}

		y, err := strconv.Atoi(fields[3])
		if err != nil {
			return fmt.Errorf("Invalid Y coordinate: %v", err)
		}

		if !g.CanPlaceShip(size, x, y, isVertical) {
			return fmt.Errorf("Cannot place ship of size %d at (%d, %d)", size, x, y)
		}

		g.PlaceShip(size, x, y, isVertical)
		g.shipCounts[size]++
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func (g *Game) HandleShotCommand(x, y int) string {
	if !g.gameStarted {
		return "failed"
	}
	if !g.IsValidCoordinate(x, y) {
		return "failed"
	}
	for _, shot := range g.shotHistory {
		if shot.X == x && shot.Y == y {
			return "failed"
		}
	}
	g.shotHistory = append(g.shotHistory, Pair{X: x, Y: y})
	if g.field[y][x] == 'S' {
		for i := range g.ships {
			ship := &g.ships[i]
			if ship.IsVertical {
				if x == ship.X && y >= ship.Y && y < ship.Y+ship.Size {
					ship.Hits++
					g.field[y][x] = 'X'
					return func() string {
						if ship.Hits == ship.Size {
							return "kill"
						}
						return "hit"
					}()
				}
			} else {
				if y == ship.Y && x >= ship.X && x < ship.X+ship.Size {
					ship.Hits++
					g.field[y][x] = 'X'
					return func() string {
						if ship.Hits == ship.Size {
							return "kill"
						}
						return "hit"
					}()
				}
			}
		}
	}
	g.field[y][x] = 'O'
	return "miss"
}

func (g *Game) ProcessShotResult(result string) {
	switch result {
	case "hit":
		g.lastShotResult = HIT
	case "kill":
		g.lastShotResult = KILL
		g.hits = nil
	case "miss":
		g.lastShotResult = MISS
	}
}

func (g *Game) GetNextShotFromHit(x, y int) Pair {
	nextShot := Pair{X: x + 1, Y: y}
	if g.IsValidCoordinate(nextShot.X, nextShot.Y) && g.field[nextShot.Y][nextShot.X] == '.' {
		return nextShot
	}
	nextShot = Pair{X: x - 1, Y: y}
	if g.IsValidCoordinate(nextShot.X, nextShot.Y) && g.field[nextShot.Y][nextShot.X] == '.' {
		return nextShot
	}
	nextShot = Pair{X: x, Y: y + 1}
	if g.IsValidCoordinate(nextShot.X, nextShot.Y) && g.field[nextShot.Y][nextShot.X] == '.' {
		return nextShot
	}
	nextShot = Pair{X: x, Y: y - 1}
	if g.IsValidCoordinate(nextShot.X, nextShot.Y) && g.field[nextShot.Y][nextShot.X] == '.' {
		return nextShot
	}
	return Pair{X: -1, Y: -1}
}

func (g *Game) ProcessHit(x, y int) {
	g.hits = append(g.hits, Pair{X: x, Y: y})
	g.lastShotResult = HIT
}

func (g *Game) GetNextShot() Pair {
	if g.strategy == "custom" {
		if len(g.hits) > 0 {
			if len(g.hits) == 2 {
				hit1 := g.hits[0]
				hit2 := g.hits[1]
				x1, y1 := hit1.X, hit1.Y
				x2, y2 := hit2.X, hit2.Y
				if x1 == x2 {
					if y1 < y2 {
						return g.GetNextShotFromHit(x1, y2+1)
					} else {
						return g.GetNextShotFromHit(x1, y1-1)
					}
				} else if y1 == y2 {
					if x1 < x2 {
						return g.GetNextShotFromHit(x2+1, y1)
					} else {
						return g.GetNextShotFromHit(x1-1, y1)
					}
				}
			}
			for _, hit := range g.hits {
				nextShot := g.GetNextShotFromHit(hit.X, hit.Y)
				if nextShot.X != -1 && nextShot.Y != -1 {
					return nextShot
				}
			}
		}
		for {
			x := rand.Intn(g.width)
			y := rand.Intn(g.height)
			if g.field[y][x] == '.' {
				return Pair{X: x, Y: y}
			}
		}
	}
	if g.strategy == "ordered" {
		if g.nextShotX == 0 && g.nextShotY == 0 {
			g.nextShotX = 0
			g.nextShotY = 0
		}
		if g.nextShotY < g.height {
			nextShot := Pair{X: g.nextShotX, Y: g.nextShotY}
			g.nextShotX++
			if g.nextShotX >= g.width {
				g.nextShotX = 0
				g.nextShotY++
			}
			return nextShot
		}
	}
	return Pair{X: -1, Y: -1}
}

func (g *Game) IsGameFinished() bool {
	for _, ship := range g.ships {
		if ship.Hits < ship.Size {
			return false
		}
	}
	return true
}

func (g *Game) IsWinner() bool {
	return g.IsGameFinished()
}

func (g *Game) IsLoser() bool {
	return g.IsGameFinished()
}

func (g *Game) HandleCommand(commandLine string) string {
	fields := strings.Fields(commandLine)
	if len(fields) == 0 {
		return "failed"
	}

	command := fields[0]
	args := fields[1:]

	switch command {
	case "ping":
		return "pong"
	case "exit":
		os.Exit(0)
	case "create":
		if len(args) < 1 {
			return "failed"
		}
		roleStr := args[0]
		g.isGameCreated = true
		switch roleStr {
		case "master":
			g.role = MASTER
			g.width = 75000
			g.height = 75000
			g.shipCounts[1] = 4
			g.shipCounts[2] = 2
			g.shipCounts[3] = 2
			g.shipCounts[4] = 1
			err := g.RandomizeShipPlacement()
			if err != nil {
				return "failed"
			}
		case "slave":
			g.role = SLAVE
		default:
			g.role = NONE
		}
		if g.role != NONE {
			return "ok"
		}
		return "failed"
	case "start":
		return g.HandleStartCommand()
	case "stop":
		if g.gameStarted {
			g.gameStarted = false
			return "ok"
		}
		return "failed"
	case "set":
		if len(args) < 2 {
			return "failed"
		}
		param := args[0]
		switch param {
		case "width":
			value, err := strconv.Atoi(args[1])
			if err != nil || value <= 0 {
				return "failed"
			}
			g.width = value
			return "ok"
		case "height":
			value, err := strconv.Atoi(args[1])
			if err != nil || value <= 0 {
				return "failed"
			}
			g.height = value
			return "ok"
		case "count":
			if len(args) < 3 {
				return "failed"
			}
			typeValue, err := strconv.Atoi(args[1])
			if err != nil || typeValue < 1 || typeValue > 4 {
				return "failed"
			}
			value, err := strconv.Atoi(args[2])
			if err != nil || value < 0 {
				return "failed"
			}
			g.shipCounts[typeValue] = value
			return "ok"
		case "strategy":
			if len(args) < 2 {
				return "failed"
			}
			strat := args[1]
			if strat == "ordered" || strat == "custom" {
				g.strategy = strat
				return "ok"
			}
			return "failed"
		case "result":
			if len(args) < 2 {
				return "failed"
			}
			result := args[1]
			if result == "miss" {
				g.lastShotResult = MISS
				return "ok"
			} else if result == "hit" {
				g.lastShotResult = HIT
				return "ok"
			} else if result == "kill" {
				g.lastShotResult = KILL
				return "ok"
			}
			return "failed"
		}
		return "failed"
	case "get":
		if len(args) < 1 {
			return "failed"
		}
		param := args[0]
		switch param {
		case "width":
			if g.width > 0 {
				return strconv.Itoa(g.width)
			}
			return "failed"
		case "height":
			if g.height > 0 {
				return strconv.Itoa(g.height)
			}
			return "failed"
		case "count":
			if len(args) < 2 {
				return "failed"
			}
			typeValue, err := strconv.Atoi(args[1])
			if err != nil || typeValue < 1 || typeValue > 4 {
				return "failed"
			}
			return strconv.Itoa(g.shipCounts[typeValue])
		case "strategy":
			return g.strategy
		}
		return "failed"
	case "shot":
		if len(args) == 0 {
			nextShot := g.GetNextShot()
			if nextShot.X >= 0 && nextShot.Y >= 0 {
				fmt.Printf("%d %d\n", nextShot.X, nextShot.Y)
				return ""
			}
			return "failed"
		} else if len(args) >= 2 {
			x, err1 := strconv.Atoi(args[0])
			y, err2 := strconv.Atoi(args[1])
			if err1 != nil || err2 != nil {
				return "failed"
			}
			return g.HandleShotCommand(x, y)
		}
		return "failed"
	case "finished":
		if g.IsGameFinished() {
			return "yes"
		}
		return "no"
	case "win":
		if g.IsWinner() {
			return "yes"
		}
		return "no"
	case "lose":
		if g.IsLoser() {
			return "yes"
		}
		return "no"
	case "dump":
		if len(args) < 1 {
			return "failed"
		}
		path := args[0]
		err := g.SaveToFile(path)
		if err != nil {
			return "failed"
		}
		return "ok"
	case "load":
		if len(args) < 1 {
			return "failed"
		}
		path := args[0]
		err := g.LoadFromFile(path)
		if err != nil {
			return "failed"
		}
		return "ok"
	}
	return "failed"
}
