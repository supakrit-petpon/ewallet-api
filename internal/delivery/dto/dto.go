package dto

import "time"


type Response struct {
    Success bool          `json:"success"`
    Code    string          `json:"code"`
    Message string          `json:"message"`
    Data    any         `json:"data,omitempty"`
    Error   *ErrorBody  `json:"error,omitempty"`
}

type TransactionData struct {
    RefID          string    `json:"ref_id,omitempty"`
    SourceID       *uint     `json:"source_id,omitempty"`
    DestinationID  *uint     `json:"destination_id,omitempty"`
    Amount         string   `json:"amount,omitempty"`
    Currency       string    `json:"currency,omitempty"`
    CurrentBalance string   `json:"current_balance,omitempty"`
    CreatedAt      time.Time `json:"created_at,omitempty"`
    Status          string `json:"status,omitempty"`
    Transaction_Type string  `json:"transaction_type,omitempty"`
}

type ErrorBody struct{
    Detail string `json:"detail,omitempty"`
    Fields map[string]string `json:"fields,omitempty"`
}

