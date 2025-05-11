package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

func ParseArgs(args []string, config *Config) int {
	if len(args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] logs_filename\n", args[0])
		return 1
	}

	for i := 1; i < len(args); i++ {
		arg := args[i]

		if strings.HasPrefix(arg, "--print") || arg == "-p" {
			config.Print = true
		} else if strings.HasPrefix(arg, "--output=") {
			value := strings.TrimPrefix(arg, "--output=")
			config.Output = value
		} else if strings.HasPrefix(arg, "--stats=") {
			value := strings.TrimPrefix(arg, "--stats=")
			stats, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Invalid stats value: %s\n", value)
				return 1
			}
			config.Stats = stats
		} else if strings.HasPrefix(arg, "--window=") {
			value := strings.TrimPrefix(arg, "--window=")
			window, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Invalid window value: %s\n", value)
				return 1
			}
			config.Window = window
		} else if strings.HasPrefix(arg, "--from=") {
			value := strings.TrimPrefix(arg, "--from=")
			from, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Invalid from value: %s\n", value)
				return 1
			}
			config.From = from
		} else if strings.HasPrefix(arg, "--to=") {
			value := strings.TrimPrefix(arg, "--to=")
			to, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Invalid to value: %s\n", value)
				return 1
			}
			config.To = to
		} else if arg == "-o" {
			if i+1 >= len(args) {
				fmt.Fprintf(os.Stderr, "Missing output value\n")
				return 1
			}
			i++
			config.Output = args[i]
		} else if arg == "-s" {
			if i+1 >= len(args) {
				fmt.Fprintf(os.Stderr, "Missing stats value\n")
				return 1
			}
			i++
			stats, err := strconv.ParseInt(args[i], 10, 64)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Invalid stats value: %s\n", args[i])
				return 1
			}
			config.Stats = stats
		} else if arg == "-w" {
			if i+1 >= len(args) {
				fmt.Fprintf(os.Stderr, "Missing window value\n")
				return 1
			}
			i++
			window, err := strconv.ParseInt(args[i], 10, 64)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Invalid window value: %s\n", args[i])
				return 1
			}
			config.Window = window
		} else if arg == "-f" {
			if i+1 >= len(args) {
				fmt.Fprintf(os.Stderr, "Missing from value\n")
				return 1
			}
			i++
			from, err := strconv.ParseInt(args[i], 10, 64)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Invalid from value: %s\n", args[i])
				return 1
			}
			config.From = from
		} else if arg == "-e" {
			if i+1 >= len(args) {
				fmt.Fprintf(os.Stderr, "Missing to value\n")
				return 1
			}
			i++
			to, err := strconv.ParseInt(args[i], 10, 64)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Invalid to value: %s\n", args[i])
				return 1
			}
			config.To = to
		} else if config.Path == "" && !strings.HasPrefix(arg, "-") {
			config.Path = arg
		} else {
			fmt.Fprintf(os.Stderr, "Unknown option: %s\n", arg)
			return 1
		}
	}

	if config.Path == "" {
		fmt.Fprintf(os.Stderr, "Missing path to log file\n")
		return 1
	}

	return 0
}

func main() {
	var config Config

	if ParseArgs(os.Args, &config) != 0 {
		os.Exit(1)
	}

	if ProcessLog(&config) != 0 {
		os.Exit(1)
	}
}
