package http

import (
	"piano/e-wallet/internal/domain"
	"piano/e-wallet/internal/usecases"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gofiber/fiber/v3"
)

type AuthHandler struct{
	authUserCase usecases.AuthUseCase
}

func NewAuthHandler(authUserCase usecases.AuthUseCase) *AuthHandler{
	return &AuthHandler{authUserCase: authUserCase}
}

func (h *AuthHandler) Login(c fiber.Ctx) error {
	var user struct {
        Email    string `json:"email" validate:"required,email"`
        Password string `json:"password" validate:"required"`
    }

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

	//Login
	token, err := h.authUserCase.Login(user.Email, user.Password)
	if err != nil {
		if strings.Contains(err.Error(), "Invalid email or password"){
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		if strings.Contains(err.Error(), "Internal server error"){
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": err.Error(),
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
	
	return c.JSON(fiber.Map{
			"message": "Login successful!",
		})
}
