package dto

type AmountRequest struct{
	Amount float64 `json:"amount" validate:"required,gte=0"`
}

type TransferRequest struct{
	DestinationID uint `json:"destination_id" validate:"required"`
	Amount float64 `json:"amount" validate:"required,gte=0"`
}