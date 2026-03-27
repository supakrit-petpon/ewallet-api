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
    ErrInsufficientBalance = errors.New("insufficient balance for this transaction")
    ErrConflictSourceDesId  = errors.New("source_id and destination_id can't be same")

    // Internal Server error
    ErrInternalServerError = errors.New("internal server error")
)
var(
    // Authentication & Authorization
    ERR_AUTH_UNAUTHORIZED    = "AUTH_UNTHORIZED"
    
    // Resource Not Found (ใช้คำว่า NotFound แทน RecordNotFound จะสั้นกว่า)
    ERR_NOT_FOUND_USER          = "NOT_FOUND_USER"
    ERR_NOT_FOUND_WALLET        = "NOT_FOUND_WALLET"
    ERR_NOT_FOUND_TRANSACTION   = "NOT_FOUND_TRANSACTION"
    
    // Validation & Logic
    ERR_INVALID_REQUEST         = "INVALID_REQUEST"
    ERR_VALIDATION_INVALID_INPUT = "VALIDATION_INVALID_INPUT"
    ERR_CONFLICT_EMAIL          = "CONFLICT_EMAIL"
    ERR_CONFLICT_USER_WALLET         = "CONFLICT_USER_WALLET"
    ERR_CONFLICT_TRANSACTION_REFID       = "CONFLICT_TRANSACTION_REFID"
    ERR_INSUFFICIENT_BALANCE      = "INSUFFICIENT_BALANCE"
    ERR_CONFLITCT_SOURCE_DES_ID = "CONFLITCT_SOURCE_DES_ID"
    
    // Internal Server error
    ERR_INTERNAL_ERROR = "INTERNAL_ERROR"
)

//For Fields Validation
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