# Basiq API Task

This project integrates with the Basiq API and calculates the average spending per transaction category.

The application performs the following steps:

1. Authenticate with the Basiq API and obtain an access token
2. Create a sandbox user
3. Create a connection to the sandbox bank (AU00000)
4. Poll the asynchronous job until the connection is completed
5. Retrieve all user transactions
6. Calculate the average spending per category (`subClass.code`)
7. Return the results as JSON via a REST endpoint

# Requirements

- Go 1.20+
- Basiq API key

# Setup

Set your API key as an environment variable:


export BASIQ_API_KEY=YOUR_API_KEY


You can obtain your API key from the Basiq Developer Dashboard.

# Run

Start the application:


go run .


The application will start an HTTP server on:


http://localhost:8080


# API Endpoint

To trigger the analysis, call:


GET http://localhost:8080/average-spending


Example response:


[
    {
        "code": "411",
        "title": "Supermarket and Grocery Stores",
        "average": 175.60
    },
    {
        "code": "422",
        "title": "Electrical and Electronic Goods Retailing",
        "average": 16.17
    }
]


# Run tests


go test ./...


# Project structure


basiq-task
├── main.go // HTTP server and endpoints
├── analysis.go // business logic for transaction analysis
├── basiq_client.go // Basiq API client
├── models.go // request and response models
└── go.mod

# Reliability

The application includes a few reliability features to make the integration with the Basiq API more robust.

## Access token caching

The Basiq API requires exchanging an API key for an access token.  
Instead of requesting a new token for every operation, the application caches the token in memory until it expires.

The token cache stores:

- the access token
- the expiration time returned by the API

When a token is requested:

1. If a valid token already exists in the cache, it is reused.
2. If the token has expired (or is about to expire), a new one is requested from the API.

A small safety buffer is applied to avoid using tokens that are about to expire.

This reduces unnecessary calls to the `/token` endpoint and improves performance.

## Retry for transient HTTP errors

The HTTP client implements a simple retry strategy for transient failures such as:

- network errors
- server errors (HTTP 5xx)
- rate limiting responses (HTTP 429)

Requests are retried up to **3 times** with a short backoff delay between attempts.

This helps the application recover from temporary network issues or API instability without failing the entire analysis.

## Connection pooling

The HTTP client instance is reused across requests, allowing Go's HTTP transport to reuse TCP connections.

This improves performance and reduces latency when communicating with the Basiq API.

# Implementation notes

- The Basiq API uses asynchronous jobs when creating bank connections.
- The application polls the job status every 2 seconds until all steps are successful.
- Transactions are retrieved using pagination until all pages are processed.
- Only **debit transactions (negative amounts)** are considered spending.
- The HTTP client is reused to allow connection pooling.

# Postman collection

A Postman collection for testing the API is included in this repository.

Location:

postman/go-task.postman_collection.json

You can import this collection into Postman and run the requests step by step to:

1. Obtain an access token
2. Create a sandbox user
3. Create a bank connection
4. Poll the job status until the connection is completed
5. Retrieve transactions
6. Call the application endpoint

Example endpoint:

GET http://localhost:8080/average-spending

This collection demonstrates both direct Basiq API calls and the final application endpoint.
