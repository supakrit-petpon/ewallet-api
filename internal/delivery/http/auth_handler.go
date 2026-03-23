package http

import (
	"errors"
	"piano/e-wallet/internal/domain"
	"piano/e-wallet/internal/usecases"
	"piano/e-wallet/pkg/logger"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

type AuthHandler struct{
	authUserCase usecases.AuthUseCase
	logger logger.Logger
}

func NewAuthHandler(authUserCase usecases.AuthUseCase, logger logger.Logger) *AuthHandler{
	return &AuthHandler{authUserCase: authUserCase, logger: logger}
}

func (h *AuthHandler) Login(c fiber.Ctx) error {
	var user struct {
        Email    string `json:"email" validate:"required,email"`
        Password string `json:"password" validate:"required"`
    }

	if err := c.Bind().Body(&user); err != nil{
		return c.Status(400).JSON(fiber.Map{
			"error": "invalid request body",
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

	//Login
	token, err := h.authUserCase.Login(user.Email, user.Password)
	if err != nil {
		switch{
		case errors.Is(err, domain.ErrAuthUnauthorized):
			return c.Status(401).JSON(fiber.Map{
				"message": err.Error(),
			})
		default:
			h.logger.Error("unexpected error in login handler", err, "path", c.Path())
			return c.Status(500).JSON(fiber.Map{
				"message": "something went wrong",
			})
		}
    }
		
	//Collet token to Cookies
	c.Cookie(&fiber.Cookie{
			Name: "jwt",
			Value: token,
			Expires: time.Now().Add(time.Hour * 3),
			HTTPOnly: true,  // ป้องกัน XSS
			Secure:   false, // พัฒนาบน localhost ให้เป็น false ก่อน ถ้าขึ้น Production ต้อง true
			SameSite: "Lax",  // ป้องกัน CSRF
		})
	
	return c.Status(200).JSON(fiber.Map{
			"message": "Login successful!",
		})
}
