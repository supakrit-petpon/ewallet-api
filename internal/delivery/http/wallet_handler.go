package http

import (
	"errors"
	"fmt"
	"piano/e-wallet/internal/delivery/dto"
	"piano/e-wallet/internal/domain"
	"piano/e-wallet/internal/usecases"
	"piano/e-wallet/pkg/logger"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

type WalletHandler struct{
	walletUseCase usecases.WalletUseCase
	logger	logger.Logger
}

func NewWalletHandler(walletUseCase usecases.WalletUseCase, logger	logger.Logger) *WalletHandler{
	return &WalletHandler{walletUseCase: walletUseCase, logger: logger}
}

func (h *WalletHandler) Balance(c fiber.Ctx) error {
	userId := c.Locals("userId").(uint)
	
	balance, err := h.walletUseCase.Balance(userId)
	if err != nil {
		switch{
		case errors.Is(err, domain.ErrNotFoundWallet):
			return c.Status(404).JSON(fiber.Map{
				"message": "wallet is not found",
			})
		default:
			h.logger.Error("unexpected error in wallet handler", err, "path", c.Path())
			return c.Status(500).JSON(fiber.Map{
				"message": "something went wrong",
			})
		}
	}

	return c.Status(200).JSON(fiber.Map{
		"balance" : balance,
	})
}

func (h *WalletHandler) TopUp(c fiber.Ctx) error{
	userId := c.Locals("userId").(uint)

	var req struct {
        Amount float64 `json:"amount" validate:"required,gte=0"`
    }

	if err := c.Bind().Body(&req); err != nil{
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid request",
		})
	}

	//Validate
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
	var errMsgs []string
	if ve, ok := err.(validator.ValidationErrors); ok{
		for _, fe := range ve{
			errMsgs = append(errMsgs, domain.GetErrorMessage(fe))
		}
		return c.Status(400).JSON(fiber.Map{
			"errors" : errMsgs,
		})
		}
	}	

	transaction, balance, err := h.walletUseCase.TopUp(userId, req.Amount)
	if err != nil {
		switch{
			case errors.Is(err, domain.ErrNotFoundWallet):
				return c.Status(404).JSON(fiber.Map{
					"message": "wallet not found",
				})
			case errors.Is(err, domain.ErrConflictTransactionRefId):
				return c.Status(409).JSON(fiber.Map{
					"message": "transaction is already created",
				})
			case errors.Is(err, domain.ErrNotFoundTransaction):
					return c.Status(404).JSON(fiber.Map{
						"message": "transaction record not found",
					})
			default:
				h.logger.Error("unexpected error in wallet handler", err, "path", c.Path())
				return c.Status(500).JSON(fiber.Map{
					"message": "something went wrong",
				})
		}
	}

	response := dto.TopUpResponse{
		Status: "success",
		Data: dto.TopUpData{
			RefID: transaction.ReferenceID,
			Amount: fmt.Sprintf("%.2f", float64(transaction.Amount)/100),
			Currency: "THB",
			CurrentBalance: fmt.Sprintf("%.2f", float64(balance)/100),
			CreatedAt: transaction.CreatedAt,
		},		
	}
	return c.Status(200).JSON(response)
}