package http

import (
	"errors"
	"piano/e-wallet/internal/usecases"
	"strings"

	"github.com/gofiber/fiber/v3"
)

type WalletHandler struct{
	walletUseCase usecases.WalletUseCase
}

func NewWalletHandler(walletUseCase usecases.WalletUseCase) *WalletHandler{
	return &WalletHandler{walletUseCase: walletUseCase}
}

func (h *WalletHandler) Balance(c fiber.Ctx) error {
	userId := c.Locals("userId").(int)
	
	balance, err := h.walletUseCase.Balance(uint(userId))
	if err != nil {
		if strings.Contains(err.Error(), "Invalid user id") || errors.Is(err, fiber.ErrNotFound){
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		if strings.Contains(err.Error(), "Internal server error") || errors.Is(err, fiber.ErrInternalServerError){
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
			})
		}
	}

	return c.JSON(fiber.Map{
		"balance" : balance,
	})
}