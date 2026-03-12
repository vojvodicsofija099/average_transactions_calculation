package main

import (
	"math"
	"testing"
)

func TestCalculateAveragesPerCategory(t *testing.T) {
	transactions := []Transaction{
		{
			Amount: "-10.00",
			SubClass: &SubClass{
				Code:  "411",
				Title: "Supermarket",
			},
		},
		{
			Amount: "-20.00",
			SubClass: &SubClass{
				Code:  "411",
				Title: "Supermarket",
			},
		},
		{
			Amount: "-30.00",
			SubClass: &SubClass{
				Code:  "422",
				Title: "Electronics",
			},
		},
		{
			Amount: "50.00", // income, should be ignored
			SubClass: &SubClass{
				Code:  "411",
				Title: "Supermarket",
			},
		},
	}

	result := calculateAveragesPerCategory(transactions)

	resultByCode := make(map[string]CategoryAverage)
	for _, categoryAverage := range result {
		resultByCode[categoryAverage.Code] = categoryAverage
	}

	supermarket, exists := resultByCode["411"]
	if !exists {
		t.Fatalf("expected category 411 to exist")
	}

	if supermarket.Title != "Supermarket" {
		t.Errorf("expected title for 411 to be Supermarket, got %s", supermarket.Title)
	}

	if math.Abs(supermarket.Average-15.0) > 0.001 {
		t.Errorf("expected average for 411 to be 15.0, got %f", supermarket.Average)
	}

	electronics, exists := resultByCode["422"]
	if !exists {
		t.Fatalf("expected category 422 to exist")
	}

	if electronics.Title != "Electronics" {
		t.Errorf("expected title for 422 to be Electronics, got %s", electronics.Title)
	}

	if math.Abs(electronics.Average-30.0) > 0.001 {
		t.Errorf("expected average for 422 to be 30.0, got %f", electronics.Average)
	}
}
