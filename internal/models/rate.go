package models

import "time"

type Rate struct {
	Base       string
	Target     string
	Value      float64
	Date       time.Time
	Source     string
	Fetched    time.Time
	Calculated bool
}

type RateRequest struct {
	Base    string
	Targets []string
	Date    time.Time
}
