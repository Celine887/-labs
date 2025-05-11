package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"./dataflow"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <directory>")
		os.Exit(1)
	}

	dirPath := os.Args[1]
	recursive := true

	pipeline := dataflow.Dir(dirPath, recursive)

	flow := dataflow.Pipe(pipeline.GetFlow(), dataflow.Filter(func(path string) bool {
		return filepath.Ext(path) == ".txt"
	}))

	flow = dataflow.Pipe(flow, dataflow.OpenFiles())

	flow = dataflow.Pipe(flow, dataflow.Transform(func(file dataflow.FileContent) string {
		return string(file.Content)
	}))

	flow = dataflow.Pipe(flow, dataflow.Split("\n ,.;"))

	flow = dataflow.Pipe(flow, dataflow.Filter(func(token string) bool {
		return token != ""
	}))

	flow = dataflow.Pipe(flow, dataflow.Transform(func(token string) string {
		return strings.ToLower(token)
	}))

	flow = dataflow.Pipe(flow, dataflow.AggregateByKey(
		0,
		func(token string, count int) int {
			return count + 1
		},
		func(token string) string {
			return token
		},
	))

	flow = dataflow.Pipe(flow, dataflow.Transform(func(kv dataflow.KV[string, int]) string {
		return fmt.Sprintf("%s - %d", kv.Key, kv.Value)
	}))

	_ = dataflow.Pipe(flow, dataflow.Out(os.Stdout))

}
