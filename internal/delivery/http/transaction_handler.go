package http

import (
	"errors"
	"piano/e-wallet/internal/delivery/dto"
	"piano/e-wallet/internal/domain"
	"piano/e-wallet/internal/usecases"
	"piano/e-wallet/pkg/logger"

	"github.com/gofiber/fiber/v3"
)

type TransactionHandler struct{
	service usecases.TransactionUseCase
	logger logger.Logger
}

func NewTransactionHandler(service usecases.TransactionUseCase, logger logger.Logger) *TransactionHandler {
	return &TransactionHandler{service: service, logger: logger}
}

// TransactionHandler
// @Summary      Get Transaction
// @Description  Get Transaction by reference_id
// @Tags		 transaction
// @Security     JWTToken
// @Accept       json
// @Produce      json
// @Param refId  path 	  string true "refId of transaction"
// @Success      200      {object}  dto.Response "OK"
// @Failure      400      {object}  dto.Response "Bad Request"
// @Failure      404      {object}  dto.Response "Not Found"
// @Failure      500      {object}  dto.Response "Internal Server Error"
// @Router       /transaction/{refId} [get]
func (h *TransactionHandler) GetTransaction(c fiber.Ctx) error {
	refId := c.Params("refId")
	if refId == "" {
		return c.Status(400).JSON(&dto.Response{
			Success: false,
			Code: domain.ERR_INVALID_REQUEST,
			Message: "refId param is missing",
		})
	}
	
	transaction, err := h.service.GetTransaction(refId)
	if err != nil {
		var status int
		var code string
		var message string

		switch {
			case errors.Is(err, domain.ErrNotFoundTransaction):
				status, code, message = 404, domain.ERR_NOT_FOUND_TRANSACTION, "transaction record not found"
			default:
				h.logger.Error("unexpected error in wallet handler", err, "path", c.Path())
				status, code, message = 500, domain.ERR_INTERNAL_ERROR, "Something went wrong"
        }
		
		resp := &dto.Response{
            Success: false,
            Code:    code,
            Message: message,
        }
		return c.Status(status).JSON(resp)
	}

	return c.JSON(&dto.Response{
		Success: true,
		Data: transaction,
	})
}

// TransactionHandler
// @Summary      Get Transaction List
// @Description  Get Transaction List
// @Tags		 transaction
// @Security     JWTToken
// @Accept       json
// @Produce      json
// @Success      200      {object}  dto.Response "OK"
// @Failure      404      {object}  dto.Response "Not Found"
// @Failure      500      {object}  dto.Response "Internal Server Error"
// @Router       /transaction [get]
func (h *TransactionHandler) GetAllTransaction(c fiber.Ctx) error {
	val := c.Locals("userId")
    userId, ok := val.(uint)
    if !ok {
        return c.Status(500).JSON(&dto.Response{
			Success: false,
			Code: domain.ERR_INTERNAL_ERROR,
			Message: "internal server error: user context missing",
		})
    }
	
	transactions, err := h.service.GetAllTransaction(userId)
	if err != nil {
		var status int
		var code string
		var message string

		switch {
			case errors.Is(err, domain.ErrNotFoundWallet):
				status, code, message = 404, domain.ERR_NOT_FOUND_WALLET, "wallet record not found"
			default:
				h.logger.Error("unexpected error in transaction handler", err, "path", c.Path())
				status, code, message = 500, domain.ERR_INTERNAL_ERROR, "Something went wrong"
        }
		
		resp := &dto.Response{
            Success: false,
            Code:    code,
            Message: message,
        }
		return c.Status(status).JSON(resp)
	}

	return c.JSON(&dto.Response{
		Success: true,
		Data: transactions,
	})
}