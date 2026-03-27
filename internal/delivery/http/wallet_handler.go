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

func NewWalletHandler(walletUseCase usecases.WalletUseCase, logger logger.Logger) *WalletHandler{
	return &WalletHandler{walletUseCase: walletUseCase, logger: logger}
}

// WalletHandler
// @Summary      Wallet Balance
// @Description  Get balance from wallet
// @Tags		 wallet
// @Security     JWTToken
// @Accept       json
// @Produce      json
// @Success      200      {object}  dto.Response "OK"
// @Failure      404      {object}  dto.Response "Not Found"
// @Failure      500      {object}  dto.Response "Internal Server Error"
// @Router       /wallet/balance [get]
func (h *WalletHandler) Balance(c fiber.Ctx) error {
	userId := c.Locals("userId").(uint)
	
	balance, err := h.walletUseCase.Balance(userId)
	if err != nil {
		var status int
		var code string
		var message string

		switch {
			case errors.Is(err, domain.ErrNotFoundWallet):
				status, code, message = 404, domain.ERR_AUTH_UNAUTHORIZED, "wallet is not exist"
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

	return c.Status(200).JSON(&dto.Response{
				Success: true,
				Data: struct{
					Balance string `json:"balance"`
				}{
					Balance: balance,
				},
			})
}

// WalletHandler
// @Summary      Wallet TopUp
// @Description  Topup to own wallet
// @Tags		 wallet
// @Security     JWTToken
// @Accept       json
// @Produce      json
// @Param        request  body      dto.AmountRequest  true "Amount Request"
// @Success      200      {object}  dto.Response "OK"
// @Failure      400      {object}  dto.Response "Bad Request"
// @Failure      404      {object}  dto.Response "Not Found"
// @Failure      500      {object}  dto.Response "Internal Server Error"
// @Router       /wallet/topup [post]
func (h *WalletHandler) TopUp(c fiber.Ctx) error{
	val := c.Locals("userId")
    userId, ok := val.(uint)
    if !ok {
        return c.Status(500).JSON(&dto.Response{
			Success: false,
			Code: domain.ERR_INTERNAL_ERROR,
			Message: "internal server error: user context missing",
		})
    }

	var req dto.AmountRequest
	//Invalid Struct
	if err := c.Bind().Body(&req); err != nil{
		return c.Status(400).JSON(&dto.Response{
			Success: false,
			Code: "INVALID_REQUEST",
			Message: "invalid request",
		})
	}
	//Validate
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
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

	transaction, balance, err := h.walletUseCase.TopUp(userId, req.Amount)
    if err != nil {
        var status int
        var code string
        var message string

        // 1. กำหนดค่าตามประเภท Error
        switch {
        case errors.Is(err, domain.ErrNotFoundWallet):
            status, code, message = 404, domain.ERR_NOT_FOUND_WALLET, "wallet not found"
        case errors.Is(err, domain.ErrConflictTransactionRefId):
            status, code, message = 409, domain.ERR_CONFLICT_TRANSACTION_REFID, "this transaction is already created"
        case errors.Is(err, domain.ErrNotFoundTransaction):
            status, code, message = 404, domain.ERR_NOT_FOUND_TRANSACTION, "transaction record not found"
        default:
            h.logger.Error("unexpected error in wallet handler", err, "path", c.Path())
            status, code, message = 500, domain.ERR_INTERNAL_ERROR, "Something went wrong"
        }

        // 2. สร้าง Response Object ครั้งเดียว
        resp := &dto.Response{
            Success: false,
            Code:    code,
            Message: message,
        }

        // 3. แนบ Data ถ้า UseCase คืนค่า transaction มาให้ (เช่น เคสที่บันทึก FAILED ลง DB แล้ว)
        if transaction != nil {
            resp.Data = &dto.TransactionData{
                RefID:           transaction.ReferenceID,
                Status:          transaction.Status,
                Transaction_Type: transaction.TransactionType,
                CreatedAt:       transaction.CreatedAt,
            }
        }
        return c.Status(status).JSON(resp)
    }

    // Success Response (200)
    return c.Status(200).JSON(dto.Response{
        Success: true,
        Data: &dto.TransactionData{
            RefID:          transaction.ReferenceID,
            DestinationID:  transaction.DestinationID,
            Amount:         fmt.Sprintf("%.2f", req.Amount),
            Currency:       "THB",
            CurrentBalance: fmt.Sprintf("%.2f", float64(balance)/100),
            CreatedAt:      transaction.CreatedAt,
        },
    })
}

// WalletHandler
// @Summary      Wallet Withdraw
// @Description  Withdraw from own wallet
// @Tags		 wallet
// @Security     JWTToken
// @Accept       json
// @Produce      json
// @Param        request  body      dto.AmountRequest  true "Amount Request"
// @Success      200      {object}  dto.Response "OK"
// @Failure      400      {object}  dto.Response "Bad Request"
// @Failure      404      {object}  dto.Response "Not Found"
// @Failure      500      {object}  dto.Response "Internal Server Error"
// @Router       /wallet/withdraw [post]
func (h *WalletHandler) Withdraw(c fiber.Ctx) error {
	val := c.Locals("userId")
    userId, ok := val.(uint)
    if !ok {
        return c.Status(500).JSON(&dto.Response{
			Success: false,
			Code: domain.ERR_INTERNAL_ERROR,
			Message: "internal server error: user context missing",
		})
    }

	var req dto.AmountRequest

	//Invalid Struct
	if err := c.Bind().Body(&req); err != nil{
		return c.Status(400).JSON(&dto.Response{
			Success: false,
			Code: "INVALID_REQUEST",
			Message: "invalid request",
		})
	}

	//Validate
	validate := validator.New()
	if err := validate.Struct(req); err != nil {
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

	transaction, balance, err := h.walletUseCase.Withdraw(userId, req.Amount)
	if err != nil {
        var status int
        var code string
        var message string

        // 1. กำหนดค่าตามประเภท Error
        switch {
        case errors.Is(err, domain.ErrInsufficientBalance):
            status, code, message = 400, domain.ERR_INSUFFICIENT_BALANCE, "insufficient balance for this transaction"
        case errors.Is(err, domain.ErrConflictTransactionRefId):
            status, code, message = 409, domain.ERR_CONFLICT_TRANSACTION_REFID, "this transaction is already created"
        case errors.Is(err, domain.ErrNotFoundTransaction):
            status, code, message = 404, domain.ERR_NOT_FOUND_TRANSACTION, "transaction record not found"
        default:
            h.logger.Error("unexpected error in wallet handler", err, "path", c.Path())
            status, code, message = 500, domain.ERR_INTERNAL_ERROR, "Something went wrong"
        }

        // 2. สร้าง Response Object ครั้งเดียว
        resp := &dto.Response{
            Success: false,
            Code:    code,
            Message: message,
        }

        // 3. แนบ Data ถ้า UseCase คืนค่า transaction มาให้ (เช่น เคสที่บันทึก FAILED ลง DB แล้ว)
        if transaction != nil {
            resp.Data = &dto.TransactionData{
                RefID:           transaction.ReferenceID,
                Status:          transaction.Status,
                Transaction_Type: transaction.TransactionType,
                CreatedAt:       transaction.CreatedAt,
            }
        }
        return c.Status(status).JSON(resp)
    }

	// Success Response (200)
    return c.Status(200).JSON(dto.Response{
        Success: true,
        Data: &dto.TransactionData{
            RefID:          transaction.ReferenceID,
            SourceID:       transaction.SourceID,
            Amount:         fmt.Sprintf("%.2f", req.Amount),
            Currency:       "THB",
            CurrentBalance: fmt.Sprintf("%.2f", float64(balance)/100),
            CreatedAt:      transaction.CreatedAt,
        },
    })
}

// WalletHandler
// @Summary      Wallet Transfer
// @Description  Transfer to other wallet
// @Tags		 wallet
// @Security     JWTToken
// @Accept       json
// @Produce      json
// @Param        request  body      dto.TransferRequest  true "Transfer Request"
// @Success      200      {object}  dto.Response "OK"
// @Failure      400      {object}  dto.Response "Bad Request"
// @Failure      404      {object}  dto.Response "Not Found"
// @Failure      409      {object}  dto.Response "Conflict"
// @Failure      500      {object}  dto.Response "Internal Server Error"
// @Router       /wallet/transfer [post]
func (h *WalletHandler) Transfer(c fiber.Ctx) error {
    val := c.Locals("userId")
    userId, ok := val.(uint)
    if !ok {
        return c.Status(500).JSON(&dto.Response{
			Success: false,
			Code: domain.ERR_INTERNAL_ERROR,
			Message: "internal server error: user context missing",
		})
    }

	var req dto.TransferRequest

	if err := c.Bind().Body(&req); err != nil{
		return c.Status(400).JSON(&dto.Response{
			Success: false,
			Code: "INVALID_REQUEST",
			Message: "invalid request",
		})
	}

    validate := validator.New()
	if err := validate.Struct(req); err != nil {
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
    
    transaction, balance, err := h.walletUseCase.Transfer(userId, req.DestinationID, req.Amount)
	if err != nil {
        var status int
        var code string
        var message string

        // 1. กำหนดค่าตามประเภท Error
        switch {
        case errors.Is(err, domain.ErrConflictSourceDesId):
            status, code, message = 409, domain.ERR_CONFLITCT_SOURCE_DES_ID, "can't transfer to own wallet"
        case errors.Is(err, domain.ErrNotFoundWallet):
            status, code, message = 404, domain.ERR_NOT_FOUND_WALLET, "destination wallet id not found"
        case errors.Is(err, domain.ErrInsufficientBalance):
            status, code, message = 400, domain.ERR_INSUFFICIENT_BALANCE, "insufficient balance for this transaction"
        case errors.Is(err, domain.ErrConflictTransactionRefId):
            status, code, message = 409, domain.ERR_CONFLICT_TRANSACTION_REFID, "this transaction is already created"
        case errors.Is(err, domain.ErrNotFoundTransaction):
            status, code, message = 404, domain.ERR_NOT_FOUND_TRANSACTION, "transaction record not found"
        default:
            h.logger.Error("unexpected error in wallet handler", err, "path", c.Path())
            status, code, message = 500, domain.ERR_INTERNAL_ERROR, "Something went wrong"
        }

        // 2. สร้าง Response Object ครั้งเดียว
        resp := &dto.Response{
            Success: false,
            Code:    code,
            Message: message,
        }

        // 3. แนบ Data ถ้า UseCase คืนค่า transaction มาให้ (เช่น เคสที่บันทึก FAILED ลง DB แล้ว)
        if transaction != nil {
            resp.Data = &dto.TransactionData{
                RefID:           transaction.ReferenceID,
                Status:          transaction.Status,
                Transaction_Type: transaction.TransactionType,
                CreatedAt:       transaction.CreatedAt,
            }
        }
        return c.Status(status).JSON(resp)
    }

    // Success Response (200)
    return c.Status(200).JSON(dto.Response{
        Success: true,
        Data: &dto.TransactionData{
            RefID:          transaction.ReferenceID,
            SourceID:       transaction.SourceID,
            DestinationID:  transaction.DestinationID,
            Amount:         fmt.Sprintf("%.2f", req.Amount),
            Currency:       "THB",
            CurrentBalance: fmt.Sprintf("%.2f", float64(balance)/100),
            CreatedAt:      transaction.CreatedAt,
        },
    })
}

// WalletHandler
// @Summary      Wallet Info
// @Description  Get own wallet id
// @Tags		 wallet
// @Security     JWTToken
// @Accept       json
// @Produce      json
// @Success      200      {object}  dto.Response "OK"
// @Failure      400      {object}  dto.Response "Bad Request"
// @Failure      404      {object}  dto.Response "Not Found"
// @Failure      500      {object}  dto.Response "Internal Server Error"
// @Router       /wallet/info [get]
func (h *WalletHandler) Info(c fiber.Ctx) error {
    val := c.Locals("userId")
    userId, ok := val.(uint)
    if !ok {
        return c.Status(500).JSON(&dto.Response{
			Success: false,
			Code: domain.ERR_INTERNAL_ERROR,
			Message: "internal server error: user context missing",
		})
    }
    walletId, err := h.walletUseCase.Info(userId)
    if err != nil {
        var status int
        var code string
        var message string

        // 1. กำหนดค่าตามประเภท Error
        switch {
        case errors.Is(err, domain.ErrNotFoundWallet):
            status, code, message = 404, domain.ERR_NOT_FOUND_WALLET, "wallet record not found"
        default:
            h.logger.Error("unexpected error in wallet handler", err, "path", c.Path())
            status, code, message = 500, domain.ERR_INTERNAL_ERROR, "Something went wrong"
        }

        // 2. สร้าง Response Object ครั้งเดียว
        resp := &dto.Response{
            Success: false,
            Code:    code,
            Message: message,
        }
        return c.Status(status).JSON(resp)
    }
    return c.Status(200).JSON(&dto.Response{
            Success: false,
            Code:    "SUCCESS",
            Data: map[string]uint{
                "walletId": walletId,
            },
        })
}