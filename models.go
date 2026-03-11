package main

type TokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
}

type CreateUserRequest struct {
	Email  string `json:"email"`
	Mobile string `json:"mobile"`
}

type UserResponse struct {
	Type    string `json:"type"`
	ID      string `json:"id"`
	Email   string `json:"email"`
	Mobile  string `json:"mobile"`
	Created string `json:"created"`
	Updated string `json:"updated"`
}

type Institution struct {
	ID string `json:"id"`
}

type CreateConnectionRequest struct {
	LoginID     string      `json:"loginId"`
	Password    string      `json:"password"`
	Institution Institution `json:"institution"`
}

type ConnectionResponse struct {
	Type  string `json:"type"`
	ID    string `json:"id"`
	Links Links  `json:"links"`
}

type Links struct {
	Self string `json:"self"`
	Next string `json:"next"`
}

type JobResponse struct {
	Type  string    `json:"type"`
	ID    string    `json:"id"`
	Steps []JobStep `json:"steps"`
}

type JobStep struct {
	Title  string `json:"title"`
	Status string `json:"status"`
}

type TransactionsResponse struct {
	Data  []Transaction `json:"data"`
	Links Links         `json:"links"`
}

type Transaction struct {
	ID       string    `json:"id"`
	Amount   string    `json:"amount"`
	SubClass *SubClass `json:"subClass"`
}

type SubClass struct {
	Code  string `json:"code"`
	Title string `json:"title"`
}

type CategoryAverage struct {
	Code    string  `json:"code"`
	Title   string  `json:"title"`
	Average float64 `json:"average"`
}
