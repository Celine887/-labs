package api

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"yandex-route-finder/cache"
	"yandex-route-finder/models"
)

const (
	YandexSchedulesAPIBaseURL = "https://api.rasp.yandex.net/v3.0"
	DefaultMaxTransfers       = 1
)

type YandexClient struct {
	APIKey      string
	BaseURL     string
	HTTPClient  *http.Client
	Cache       cache.Cache
	CityCodeMap map[string]string
}

func NewYandexClient(apiKey string, cache cache.Cache) *YandexClient {
	client := &YandexClient{
		APIKey:  apiKey,
		BaseURL: YandexSchedulesAPIBaseURL,
		HTTPClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		Cache:       cache,
		CityCodeMap: make(map[string]string),
	}

	client.CityCodeMap = map[string]string{
		"санкт-петербург": "c2",
		"москва":          "c213",
		"псков":           "c25",
		"новгород":        "c974",
		"выборг":          "c133",
	}

	return client
}

func (c *YandexClient) GetCityCode(cityName string) (string, error) {

	code, exists := c.CityCodeMap[cityName]
	if exists {
		return code, nil
	}

	return "", fmt.Errorf("city code for %s not found", cityName)
}

func (c *YandexClient) SearchRoutes(request models.RouteRequest) ([]models.CompleteRoute, error) {

	if routes, found := c.Cache.GetRoute(request); found {
		return routes, nil
	}

	fromCode, err := c.GetCityCode(request.FromCity)
	if err != nil {
		return nil, fmt.Errorf("failed to get code for origin city: %w", err)
	}

	toCode, err := c.GetCityCode(request.ToCity)
	if err != nil {
		return nil, fmt.Errorf("failed to get code for destination city: %w", err)
	}

	maxTransfers := request.MaxTransfers
	if maxTransfers <= 0 {
		maxTransfers = DefaultMaxTransfers
	}

	directRoutes, err := c.fetchRoutes(fromCode, toCode, request.Date, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch direct routes: %w", err)
	}

	var transferRoutes []models.CompleteRoute
	if maxTransfers > 0 {

		intermediateCity := "москва"
		intermediateCode, err := c.GetCityCode(intermediateCity)
		if err == nil {

			firstLegRoutes, err := c.fetchRoutes(fromCode, intermediateCode, request.Date, 0)
			if err == nil && len(firstLegRoutes) > 0 {
				for _, firstLeg := range firstLegRoutes {

					if len(firstLeg.Segments) > 0 {
						arrivalTime := firstLeg.Segments[len(firstLeg.Segments)-1].ArrivalTime

						transferTime := arrivalTime.Add(1 * time.Hour)

						secondLegRoutes, err := c.fetchRoutes(intermediateCode, toCode, transferTime, 0)
						if err == nil {
							for _, secondLeg := range secondLegRoutes {

								combinedRoute := models.CompleteRoute{
									Segments:      append(firstLeg.Segments, secondLeg.Segments...),
									TotalDuration: firstLeg.TotalDuration + secondLeg.TotalDuration,
									TotalPrice:    firstLeg.TotalPrice + secondLeg.TotalPrice,
									TransferCount: 1,
								}
								transferRoutes = append(transferRoutes, combinedRoute)
							}
						}
					}
				}
			}
		}
	}

	allRoutes := append(directRoutes, transferRoutes...)

	if err := c.Cache.SetRoute(request, allRoutes); err != nil {

		fmt.Printf("Warning: failed to cache routes: %v\n", err)
	}

	return allRoutes, nil
}

func (c *YandexClient) fetchRoutes(fromCode, toCode string, date time.Time, transfers int) ([]models.CompleteRoute, error) {

	cacheKey := fmt.Sprintf("route_%s_%s_%s_%d", fromCode, toCode, date.Format("2006-01-02"), transfers)
	if cachedData, found := c.Cache.Get(cacheKey); found {
		var routes []models.CompleteRoute
		if err := json.Unmarshal(cachedData, &routes); err == nil {
			return routes, nil
		}
	}

	apiURL := fmt.Sprintf("%s/search/", c.BaseURL)

	query := url.Values{}
	query.Add("apikey", c.APIKey)
	query.Add("format", "json")
	query.Add("from", fromCode)
	query.Add("to", toCode)
	query.Add("lang", "ru_RU")
	query.Add("date", date.Format("2006-01-02"))
	query.Add("transfers", fmt.Sprintf("%d", transfers))

	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.URL.RawQuery = query.Encode()

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API error: %s (status code: %d)", string(body), resp.StatusCode)
	}

	var apiResponse models.YandexAPIResponse
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		return nil, fmt.Errorf("failed to parse API response: %w", err)
	}

	routes := make([]models.CompleteRoute, 0)
	for _, segment := range apiResponse.Segments {

		departureTime, err := time.Parse("2006-01-02T15:04:05Z07:00", segment.Departure)
		if err != nil {

			departureTime, err = time.Parse("2006-01-02T15:04:05", segment.Departure)
			if err != nil {
				continue
			}
		}

		arrivalTime, err := time.Parse("2006-01-02T15:04:05Z07:00", segment.Arrival)
		if err != nil {

			arrivalTime, err = time.Parse("2006-01-02T15:04:05", segment.Arrival)
			if err != nil {
				continue
			}
		}

		routeSegment := models.Segment{
			From:             segment.From.Title,
			To:               segment.To.Title,
			DepartureTime:    departureTime,
			ArrivalTime:      arrivalTime,
			TransportType:    segment.Thread.TransportType,
			ThreadUID:        segment.Thread.UID,
			CarrierName:      segment.Thread.Carrier.Title,
			Number:           segment.Thread.Number,
			Title:            segment.Thread.Title,
			DepartureStation: segment.From.Station,
			ArrivalStation:   segment.To.Station,
			DurationMinutes:  segment.Duration,
		}

		route := models.CompleteRoute{
			Segments:      []models.Segment{routeSegment},
			TotalDuration: segment.Duration,
			TotalPrice:    0,
			TransferCount: segment.Transfers,
		}

		routes = append(routes, route)
	}

	routesData, err := json.Marshal(routes)
	if err == nil {
		if err := c.Cache.Set(cacheKey, routesData); err != nil {

			fmt.Printf("Warning: failed to cache routes data: %v\n", err)
		}
	}

	return routes, nil
}
