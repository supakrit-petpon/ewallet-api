package http

import (
	"errors"
	"piano/e-wallet/internal/delivery/dto"
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

// UserHandler
// @Summary      User Register
// @Description  Create user and return message
// @Tags		 auth
// @Accept       json
// @Produce      json
// @Param        request  body      dto.UserForRequest  true "User Information"
// @Success      201      {object}  dto.Response "Created"
// @Failure      400      {object}  dto.Response "Bad Request"
// @Failure      409      {object}  dto.Response "Conflict"
// @Failure      500      {object}  dto.Response "Internal Server Error"
// @Router       /register [post]
func (h *UserHandler) Register(c fiber.Ctx) error {
	var user domain.User
	//Invalid Struct
	if err := c.Bind().Body(&user); err != nil{
		return c.Status(400).JSON(&dto.Response{
			Success: false,
			Code: "INVALID_REQUEST",
			Message: "invalid request",
		})
	}

	//Validate
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

	err := h.userUserCase.Register(user)
	if err != nil {
		var status int
		var code string
		var message string

		switch {
			case errors.Is(err, domain.ErrConflictEmail):
				status, code, message = 409, domain.ERR_CONFLICT_EMAIL, "this email is already registered"
			case errors.Is(err, domain.ErrConflictUserWallet):
				status, code, message = 409, domain.ERR_CONFLICT_USER_WALLET, "this user is already has wallet"
			default:
				h.logger.Error("unexpected error in user handler", err, "path", c.Path())
				status, code, message = 500, domain.ERR_INTERNAL_ERROR, "Something went wrong"
        }
		
		resp := &dto.Response{
            Success: false,
            Code:    code,
            Message: message,
        }
		return c.Status(status).JSON(resp)
	}	

	return c.Status(201).JSON(&dto.Response{
				Success: true,
				Code: "SUCCESS",
				Message: "Registraion successful",
			})
}