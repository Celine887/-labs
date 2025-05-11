package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"yandex-route-finder/api"
	"yandex-route-finder/cache"
	"yandex-route-finder/models"
)

const (
	DefaultFromCity = "санкт-петербург"
	DefaultToCity   = "псков"
	CacheDir        = "./cache"
	CacheTTL        = 24 * time.Hour
)

func main() {

	apiKey := flag.String("apikey", "", "Yandex Schedules API key")
	fromCity := flag.String("from", DefaultFromCity, "Origin city")
	toCity := flag.String("to", DefaultToCity, "Destination city")
	dateStr := flag.String("date", time.Now().Format("2006-01-02"), "Date for search (YYYY-MM-DD)")
	maxTransfers := flag.Int("transfers", 1, "Maximum number of transfers")
	round := flag.Bool("round", true, "Search for round-trip routes")
	flag.Parse()

	if *apiKey == "" {
		fmt.Println("API key is required. Please set it using the -apikey flag.")
		os.Exit(1)
	}

	date, err := time.Parse("2006-01-02", *dateStr)
	if err != nil {
		fmt.Printf("Invalid date format: %s. Please use YYYY-MM-DD format.\n", *dateStr)
		os.Exit(1)
	}

	cacheDir := filepath.Join(CacheDir, "yandex_api")
	memCache, err := cache.NewMemoryCache(cacheDir, CacheTTL)
	if err != nil {
		fmt.Printf("Failed to initialize cache: %v\n", err)
		os.Exit(1)
	}

	yandexClient := api.NewYandexClient(*apiKey, memCache)

	fromCityLower := strings.ToLower(*fromCity)
	toCityLower := strings.ToLower(*toCity)

	fmt.Printf("Searching for routes from %s to %s on %s...\n",
		*fromCity, *toCity, date.Format("2006-01-02"))

	request := models.RouteRequest{
		FromCity:     fromCityLower,
		ToCity:       toCityLower,
		Date:         date,
		MaxTransfers: *maxTransfers,
	}

	routes, err := yandexClient.SearchRoutes(request)
	if err != nil {
		fmt.Printf("Error searching for routes: %v\n", err)
		os.Exit(1)
	}

	if len(routes) == 0 {
		fmt.Printf("No routes found from %s to %s on %s\n",
			*fromCity, *toCity, date.Format("2006-01-02"))
	} else {
		fmt.Printf("Found %d routes from %s to %s on %s:\n\n",
			len(routes), *fromCity, *toCity, date.Format("2006-01-02"))

		for i, route := range routes {
			fmt.Printf("Route %d:\n%s\n\n", i+1, route.Format())
		}
	}

	if *round {

		returnDate := date.AddDate(0, 0, 1)

		fmt.Printf("Searching for return routes from %s to %s on %s...\n",
			*toCity, *fromCity, returnDate.Format("2006-01-02"))

		returnRequest := models.RouteRequest{
			FromCity:     toCityLower,
			ToCity:       fromCityLower,
			Date:         returnDate,
			MaxTransfers: *maxTransfers,
		}

		returnRoutes, err := yandexClient.SearchRoutes(returnRequest)
		if err != nil {
			fmt.Printf("Error searching for return routes: %v\n", err)
			os.Exit(1)
		}

		if len(returnRoutes) == 0 {
			fmt.Printf("No return routes found from %s to %s on %s\n",
				*toCity, *fromCity, returnDate.Format("2006-01-02"))
		} else {
			fmt.Printf("Found %d return routes from %s to %s on %s:\n\n",
				len(returnRoutes), *toCity, *fromCity, returnDate.Format("2006-01-02"))

			for i, route := range returnRoutes {
				fmt.Printf("Return Route %d:\n%s\n\n", i+1, route.Format())
			}
		}
	}
}
