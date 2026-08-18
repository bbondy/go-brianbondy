package data

import (
	"testing"
	"time"
)

func TestNormalizeStravaActivityTypeGroupsWorkoutWithStairStepper(t *testing.T) {
	if got := normalizeStravaActivityType(StravaRun{Type: "Workout"}); got != "StairStepper" {
		t.Fatalf("normalizeStravaActivityType() = %q, want %q", got, "StairStepper")
	}
	if got := normalizeStravaActivityType(StravaRun{Type: "StairStepper"}); got != "StairStepper" {
		t.Fatalf("normalizeStravaActivityType() = %q, want %q", got, "StairStepper")
	}
}

func TestNormalizeStravaActivityTypeLabelsEveningRowAsRowingMachine(t *testing.T) {
	if got := normalizeStravaActivityType(StravaRun{Title: "Evening Row", Type: "WaterSport"}); got != "RowingMachine" {
		t.Fatalf("normalizeStravaActivityType() = %q, want %q", got, "RowingMachine")
	}
}

func TestRunMetricFilterMatchesDayTotals(t *testing.T) {
	minKm := 20.0
	maxKm := 40.0
	minElevation := 500
	filter := RunMetricFilter{MinDistanceKm: &minKm, MaxDistanceKm: &maxKm, MinElevationM: &minElevation}
	if filter.IsZero() {
		t.Fatal("IsZero() = true, want false for a filter with bounds")
	}
	cases := []struct {
		name       string
		distanceKm float64
		minutes    int
		elevationM int
		want       bool
	}{
		{"inside every bound", 25, 180, 900, true},
		{"on the bounds", 20, 0, 500, true},
		{"below the distance floor", 19.9, 180, 900, false},
		{"above the distance ceiling", 40.1, 180, 900, false},
		{"below the elevation floor", 25, 180, 499, false},
	}
	for _, tc := range cases {
		if got := filter.MatchesDayTotals(tc.distanceKm, tc.minutes, tc.elevationM); got != tc.want {
			t.Errorf("%s: MatchesDayTotals(%v, %d, %d) = %v, want %v", tc.name, tc.distanceKm, tc.minutes, tc.elevationM, got, tc.want)
		}
	}

	var unbounded RunMetricFilter
	if !unbounded.IsZero() {
		t.Fatal("IsZero() = false, want true for a filter without bounds")
	}
	if !unbounded.MatchesDayTotals(0, 0, 0) {
		t.Error("MatchesDayTotals() = false, want true when no bound is set")
	}
}

func TestRunInYearView(t *testing.T) {
	now := time.Date(2026, 8, 18, 0, 0, 0, 0, time.UTC)
	cases := []struct {
		name       string
		date       string
		yearFilter string
		want       bool
	}{
		{"all time keeps everything", "2019-01-01", "all", true},
		{"empty filter keeps everything", "2019-01-01", "", true},
		{"within the last 365 days", "2026-01-05", "365", true},
		{"older than 365 days", "2025-01-05", "365", false},
		{"matching year", "2024-06-01", "2024", true},
		{"other year", "2023-06-01", "2024", false},
	}
	for _, tc := range cases {
		if got := runInYearView(StravaRun{Date: tc.date}, tc.yearFilter, now); got != tc.want {
			t.Errorf("%s: runInYearView(%q, %q) = %v, want %v", tc.name, tc.date, tc.yearFilter, got, tc.want)
		}
	}
}
