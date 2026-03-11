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
				Code: "411",
			},
		},
	}

	result := calculateAveragesPerCategory(transactions)

	if math.Abs(result["411"]-(-15.0)) > 0.001 {
		t.Errorf("expected average for 411 to be -15.0, got %f", result["411"])
	}

	if math.Abs(result["422"]-(-30.0)) > 0.001 {
		t.Errorf("expected average for 422 to be -30.0, got %f", result["422"])
	}
}
