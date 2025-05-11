package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

func main() {
	game := NewGame()
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		command, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			continue
		}

		command = strings.TrimSpace(command)
		result := game.HandleCommand(command)
		if result != "" {
			fmt.Println(result)
		}
	}
}
