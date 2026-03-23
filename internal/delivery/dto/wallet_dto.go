package dto

import "time"


type TopUpResponse struct {
    Status string    `json:"status"`
    Data   TopUpData `json:"data"`
}

type TopUpData struct {
    RefID  string    `json:"ref_id"`
    Amount         string   `json:"amount"`
    Currency       string    `json:"currency"`
    CurrentBalance string   `json:"current_balance"`
    CreatedAt      time.Time `json:"created_at"`
}