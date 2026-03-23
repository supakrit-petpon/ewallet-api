package http

import (
	"errors"
	"piano/e-wallet/internal/domain"
	"piano/e-wallet/internal/usecases"
	"piano/e-wallet/pkg/logger"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)


type UserHandler struct{
	userUserCase usecases.UserUseCase
	logger 		 logger.Logger
}

func NewUserHandler(userUserCase usecases.UserUseCase, logger logger.Logger) *UserHandler{
	return &UserHandler{userUserCase: userUserCase, logger: logger}
}

func (h *UserHandler) Register(c fiber.Ctx) error {
	var user domain.User
	//Invalid Struct
	if err := c.Bind().Body(&user); err != nil{
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid request",
		})
	}

	//Validate
	validate := validator.New()
	if err := validate.Struct(user); err != nil {
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

	if err := h.userUserCase.Register(user); err != nil{
		switch{
		case errors.Is(err, domain.ErrConflictEmail):
			return c.Status(409).JSON(fiber.Map{
				"message": "this email is already registered",
			})
		case errors.Is(err, domain.ErrConflictUserWallet):
			return c.Status(409).JSON(fiber.Map{
				"message": "this user is already has wallet",
			})
		default:
			h.logger.Error("unexpected error in user handler", err, "path", c.Path())
			return c.Status(500).JSON(fiber.Map{
				"message": "something went wrong",
			})
		}
	}

	return c.Status(201).JSON(fiber.Map{
		"message": "registation success!",
	})
}

