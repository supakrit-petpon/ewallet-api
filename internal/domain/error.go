package domain

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

var (
    ErrInvalidCredentials = errors.New("Invalid email or password")
    ErrInternalServerError = errors.New("Internal server error")
    ErrWalletRecordNotFound = errors.New("Invalid wallet id")
    ErrUserRecordNotFound = errors.New("Invalid user id")
    ErrParamsFormat = errors.New("Invalid params format")
)

func GetErrorMessage(fe validator.FieldError) string {
    switch fe.Tag() {
    case "required":
        return fmt.Sprintf("%s is required", fe.Field())
    case "email":
        return "invalid email format"
    case "min":
        return fmt.Sprintf("%s is must be %s character", fe.Field(), fe.Param())
    }
    return "ข้อมูลไม่ถูกต้อง"
}