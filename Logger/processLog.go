package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Path   string
	Output string
	Print  bool
	Stats  int64
	Window int64
	From   int64
	To     int64
}

type LogEntry struct {
	TimeStr    string
	Timezone   string
	TimeStruct time.Time
	Request    string
	Status     string
}

type Request struct {
	Req   string
	Count int
}

func ParseTime(dateTime, timezone string) (time.Time, error) {

	var tzOffset int
	if len(timezone) == 5 {
		hours, err := strconv.Atoi(timezone[0:3])
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid timezone format: %s", timezone)
		}
		minutes, err := strconv.Atoi(timezone[3:5])
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid timezone format: %s", timezone)
		}
		tzOffset = (hours * 3600) + (minutes * 60)
	} else {
		return time.Time{}, fmt.Errorf("invalid timezone format: %s", timezone)
	}

	t, err := time.Parse("02/Jan/2006:15:04:05", dateTime)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date/time format: %s", dateTime)
	}

	return t.Add(time.Duration(-tzOffset) * time.Second), nil
}

func GetRequestData(str string) string {
	start := strings.Index(str, "\"")
	stop := strings.LastIndex(str, "\"")
	if start == -1 || stop == -1 {
		return ""
	}
	return str[start+1 : stop]
}

func GetTimeData(str string) string {
	start := strings.Index(str, "[")
	stop := strings.Index(str, "]")
	if start == -1 || stop == -1 {
		return ""
	}
	return str[start+1 : stop]
}

func GetStatusCode(str string) string {
	fields := strings.Fields(str)
	if len(fields) < 2 {
		return ""
	}

	quoteCount := 0
	for i, field := range fields {
		if strings.Contains(field, "\"") {
			quoteCount += strings.Count(field, "\"")
		}
		if quoteCount >= 2 && i+1 < len(fields) {

			return fields[i+1]
		}
	}
	return ""
}

func ProcessLog(config *Config) int {
	logfile, err := os.Open(config.Path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening log file: %v\n", err)
		return 1
	}
	defer logfile.Close()

	var outfile *os.File
	if config.Output != "" {
		outfile, err = os.Create(config.Output)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error opening output file: %v\n", err)
			return 1
		}
		defer outfile.Close()
	}

	scanner := bufio.NewScanner(logfile)
	requestMap := make(map[string]*Request)
	var windowVector []LogEntry

	maxWindowSize := 0
	var windowStart, windowEnd int64
	var windowStartStr, windowEndStr string

	fmt.Println("Processing log file...")

	if config.Print {
		fmt.Println("Requests with 5XX status codes:")
	}

	for scanner.Scan() {
		line := scanner.Text()
		var entry LogEntry

		timeData := GetTimeData(line)
		if timeData == "" || !strings.Contains(timeData, " ") {
			fmt.Fprintf(os.Stderr, "WARNING: Invalid time data format, skipping: %s\n", line)
			continue
		}

		parts := strings.SplitN(timeData, " ", 2)
		entry.TimeStr = parts[0]
		entry.Timezone = parts[1]

		timeStruct, err := ParseTime(entry.TimeStr, entry.Timezone)
		if err != nil {
			fmt.Fprintf(os.Stderr, "WARNING: Error parsing time, skipping: %v in line: %s\n", err, line)
			continue
		}
		entry.TimeStruct = timeStruct

		if config.From > 0 && entry.TimeStruct.Unix() < config.From {
			continue
		} else if config.To > 0 && entry.TimeStruct.Unix() > config.To {
			continue
		}

		entry.Request = GetRequestData(line)
		if entry.Request == "" {
			fmt.Fprintf(os.Stderr, "WARNING: Invalid request data format, skipping: %s\n", line)
			continue
		}

		entry.Status = GetStatusCode(line)
		if entry.Status == "" {
			fmt.Fprintf(os.Stderr, "WARNING: Invalid status code format, skipping: %s\n", line)
			continue
		}

		statusInt, err := strconv.Atoi(entry.Status)
		if err != nil || statusInt < 100 || statusInt > 599 {
			fmt.Fprintf(os.Stderr, "WARNING: Invalid status code format, skipping: %s\n", line)
			continue
		}

		if entry.Status[0] == '5' {
			if config.Output != "" {
				fmt.Fprintln(outfile, line)
			}

			if config.Print {
				fmt.Println(line)
			}

			if config.Stats > 0 {
				if req, exists := requestMap[entry.Request]; exists {
					req.Count++
				} else {
					requestMap[entry.Request] = &Request{Req: entry.Request, Count: 1}
				}
			}
		}

		if config.Window > 0 {
			windowVector = append(windowVector, entry)

			for len(windowVector) > 0 &&
				entry.TimeStruct.Unix()-windowVector[0].TimeStruct.Unix()+1 > config.Window {
				windowVector = windowVector[1:]
			}

			if len(windowVector) > maxWindowSize {
				maxWindowSize = len(windowVector)
				windowStart = windowVector[0].TimeStruct.Unix()
				windowStartStr = windowVector[0].TimeStr + " " + windowVector[0].Timezone
				windowEnd = windowVector[len(windowVector)-1].TimeStruct.Unix()
				windowEndStr = windowVector[len(windowVector)-1].TimeStr + " " + windowVector[len(windowVector)-1].Timezone
			}
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "Error reading log file: %v\n", err)
		return 1
	}

	if config.Stats > 0 {
		var requests []Request
		for _, req := range requestMap {
			requests = append(requests, *req)
		}

		sort.Slice(requests, func(i, j int) bool {
			return requests[i].Count > requests[j].Count
		})

		fmt.Println("\n==============================================================")
		fmt.Printf("%d most frequent requests with status code 5XX by occurrencies:\n", config.Stats)

		limit := int(config.Stats)
		if limit > len(requests) {
			limit = len(requests)
		}
		for i := 0; i < limit; i++ {
			fmt.Printf("%d. [%d]\t%s\n", i+1, requests[i].Count, requests[i].Req)
		}
	}

	if config.Window > 0 {
		fmt.Println("\n======================")
		fmt.Printf("Max window size: %d requests in %d seconds.\n", maxWindowSize, config.Window)
		fmt.Printf("Start: %d, %s\n", windowStart, windowStartStr)
		fmt.Printf("End:   %d, %s\n", windowEnd, windowEndStr)
	}

	return 0
}
