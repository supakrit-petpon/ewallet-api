package dto

type UserForRequest struct{
	Email string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}