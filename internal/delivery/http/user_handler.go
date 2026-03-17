package http

import (
	"piano/e-wallet/internal/domain"
	"piano/e-wallet/internal/usecases"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)


type UserHandler struct{
	userUserCase usecases.UserUseCase
}

func NewUserHandler(userUserCase usecases.UserUseCase) *UserHandler{
	return &UserHandler{userUserCase: userUserCase}
}

func (u *UserHandler) Register(c fiber.Ctx) error {
	var user domain.User
	//Invalid Struct
	if err := c.Bind().Body(&user); err != nil{
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"errors" : errMsgs,
		})
	}
}

	//Service error eg: Email already exist.
	if err := u.userUserCase.Register(user); err != nil{
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return c.SendStatus(fiber.StatusCreated)
}

