package domain

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

var (
    // Authentication & Authorization
    ErrAuthUnauthorized      = errors.New("invalid email or password")
    
    // Resource Not Found (ใช้คำว่า NotFound แทน RecordNotFound จะสั้นกว่า)
    ErrNotFoundUser         = errors.New("user record not found")
    ErrNotFoundWallet       = errors.New("wallet record not found")
    ErrNotFoundTransaction = errors.New("transaction record not found")
    
    // Validation & Logic
    ErrInvalidRequest         = errors.New("invalid request")
    ErrValidationInvalidInput = errors.New("provided input is invalid")
    ErrConflictEmail          = errors.New("this email is already registered")
    ErrConflictUserWallet = errors.New("this user is already has wallet")
    ErrConflictTransactionRefId = errors.New("this transaction is already created")

    // Internal Server error
    ErrInternalServerError = errors.New("internal server error")
)

func GetErrorMessage(fe validator.FieldError) string {
    switch fe.Tag() {
    case "required":
        return fmt.Sprintf("%s is required", fe.Field())
    case "email":
        return "invalid email format"
    case "min":
        return fmt.Sprintf("%s is must be %s character", fe.Field(), fe.Param())
    case "gte=0":
        return fmt.Sprintf("%s is must be more than 0", fe.Field())
    }
    return "ข้อมูลไม่ถูกต้อง"
}