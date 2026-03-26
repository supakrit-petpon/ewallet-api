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