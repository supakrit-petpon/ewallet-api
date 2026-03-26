package http

import (
	"errors"
	"piano/e-wallet/internal/delivery/dto"
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
		return c.Status(400).JSON(&dto.Response{
			Success: false,
			Code: "INVALID_REQUEST",
			Message: "invalid request",
		})
	}

	//Validation Fields
	validate := validator.New()
	if err := validate.Struct(user); err != nil {
		fields := map[string]string{}

		if ve, ok := err.(validator.ValidationErrors); ok {
			for _, fe := range ve {
				fields[fe.Field()] = domain.GetErrorMessage(fe)
			}
		}

		return c.Status(400).JSON(dto.Response{
			Success: false,
			Code:    "VALIDATION_ERROR",
			Message: "Invalid input",
			Error: &dto.ErrorBody{
				Fields: fields,
			},
		})
	}

	//Login
	token, err := h.authUserCase.Login(user.Email, user.Password)
	if err != nil {
		var status int
		var code string
		var message string

		switch {
			case errors.Is(err, domain.ErrAuthUnauthorized):
				status, code, message = 401, domain.ERR_AUTH_UNTHORIZED, "Invalid email or password"
			default:
				h.logger.Error("unexpected error in auth handler", err, "path", c.Path())
				status, code, message = 500, domain.ERR_INTERNAL_ERROR, "Something went wrong"
        }
		
		resp := &dto.Response{
            Success: false,
            Code:    code,
            Message: message,
        }
		return c.Status(status).JSON(resp)
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
	
	return c.Status(200).JSON(&dto.Response{
				Success: true,
				Code: "SUCCESS",
				Message: "Login successful",
			})
}
