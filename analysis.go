package main

import (
	"fmt"
	"os"
	"time"
)

func runAnalysis() ([]CategoryAverage, error) {
	apiKey := os.Getenv("BASIQ_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("BASIQ_API_KEY not set")
	}

	token, err := getAccessToken(apiKey)
	if err != nil {
		return nil, fmt.Errorf("error while getting token: %w", err)
	}

	client := NewBasiqClient(token.AccessToken)

	userRequest := CreateUserRequest{
		Email:  fmt.Sprintf("sofija-%d@example.com", time.Now().Unix()),
		Mobile: "61400000000",
	}

	user, err := client.createUser(userRequest)
	if err != nil {
		return nil, fmt.Errorf("error while creating user: %w", err)
	}

	connectionRequest := CreateConnectionRequest{
		LoginID:  "gavinBelson",
		Password: "hooli2016",
		Institution: Institution{
			ID: "AU00000",
		},
	}

	connection, err := client.createConnection(user.ID, connectionRequest)
	if err != nil {
		return nil, fmt.Errorf("error while creating connection: %w", err)
	}

	_, err = client.waitForJobSuccess(connection.ID)
	if err != nil {
		return nil, fmt.Errorf("error while waiting for job success: %w", err)
	}

	transactions, err := client.getTransactions(user.ID)
	if err != nil {
		return nil, fmt.Errorf("error getting transactions: %w", err)
	}

	averages := calculateAveragesPerCategory(transactions)
	return averages, nil
}
