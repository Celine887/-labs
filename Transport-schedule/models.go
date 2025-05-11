package models

import (
	"fmt"
	"time"
)

type RouteRequest struct {
	FromCity     string
	ToCity       string
	Date         time.Time
	MaxTransfers int
}

type Route struct {
	From            string    `json:"from"`
	To              string    `json:"to"`
	DepartureTime   time.Time `json:"departure"`
	ArrivalTime     time.Time `json:"arrival"`
	TransportType   string    `json:"transport_type"`
	CarrierName     string    `json:"carrier"`
	Price           float64   `json:"price"`
	DurationMinutes int       `json:"duration_minutes"`
}

type Segment struct {
	From             string    `json:"from"`
	To               string    `json:"to"`
	DepartureTime    time.Time `json:"departure_time"`
	ArrivalTime      time.Time `json:"arrival_time"`
	TransportType    string    `json:"transport_type"`
	ThreadUID        string    `json:"thread_uid"`
	CarrierName      string    `json:"carrier"`
	Number           string    `json:"number"`
	Title            string    `json:"title"`
	DepartureStation string    `json:"departure_station"`
	ArrivalStation   string    `json:"arrival_station"`
	DurationMinutes  int       `json:"duration_minutes"`
}

type CompleteRoute struct {
	Segments      []Segment `json:"segments"`
	TotalDuration int       `json:"total_duration_minutes"`
	TotalPrice    float64   `json:"total_price"`
	TransferCount int       `json:"transfer_count"`
}

func (r *CompleteRoute) Format() string {
	if len(r.Segments) == 0 {
		return "No route available"
	}

	result := fmt.Sprintf("Route with %d transfers (Total duration: %d minutes, Price: %.2f)\n",
		r.TransferCount, r.TotalDuration, r.TotalPrice)

	for i, seg := range r.Segments {
		result += fmt.Sprintf("Segment %d: %s → %s\n", i+1, seg.From, seg.To)
		result += fmt.Sprintf("  Transport: %s (%s)\n", seg.TransportType, seg.Title)
		result += fmt.Sprintf("  Departure: %s from %s\n", seg.DepartureTime.Format("2006-01-02 15:04"), seg.DepartureStation)
		result += fmt.Sprintf("  Arrival: %s at %s\n", seg.ArrivalTime.Format("2006-01-02 15:04"), seg.ArrivalStation)
		result += fmt.Sprintf("  Duration: %d minutes\n", seg.DurationMinutes)

		if i < len(r.Segments)-1 {
			result += fmt.Sprintf("  Transfer time: %d minutes\n",
				int(r.Segments[i+1].DepartureTime.Sub(seg.ArrivalTime).Minutes()))
		}
	}

	return result
}

type YandexAPIResponse struct {
	Segments []YandexSegment `json:"segments"`
	Search   YandexSearch    `json:"search"`
}

type YandexSearch struct {
	From struct {
		Code  string `json:"code"`
		Title string `json:"title"`
	} `json:"from"`
	To struct {
		Code  string `json:"code"`
		Title string `json:"title"`
	} `json:"to"`
	Date string `json:"date"`
}

type YandexSegment struct {
	Thread struct {
		UID     string `json:"uid"`
		Title   string `json:"title"`
		Number  string `json:"number"`
		Carrier struct {
			Code  string `json:"code"`
			Title string `json:"title"`
		} `json:"carrier"`
		TransportType string `json:"transport_type"`
		Vehicle       string `json:"vehicle"`
	} `json:"thread"`
	From struct {
		Code    string `json:"code"`
		Title   string `json:"title"`
		Station string `json:"station"`
	} `json:"from"`
	To struct {
		Code    string `json:"code"`
		Title   string `json:"title"`
		Station string `json:"station"`
	} `json:"to"`
	Departure string `json:"departure"`
	Arrival   string `json:"arrival"`
	Duration  int    `json:"duration"`
	Transfers int    `json:"transfers"`
}
