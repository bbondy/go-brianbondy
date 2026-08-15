package data

import "testing"

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
