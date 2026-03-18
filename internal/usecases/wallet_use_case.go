package usecases

import (
	"fmt"
	"piano/e-wallet/internal/domain"
)

type WalletUseCase interface{
	Balance(userId uint) (string, error)
}

type WalletService struct{
	repo domain.WalletRepository
}

func NewWalletService(repo domain.WalletRepository) WalletUseCase{
	return &WalletService{repo: repo}
}

func (s *WalletService) Balance(userId uint) (string, error){
	balance, err := s.repo.GetBalance(userId)
	if err != nil {
		return "", err
	}

	//Format balance
	formattedBalance := fmt.Sprintf("%.2f THB", float64(balance)/100)

	return formattedBalance, nil
}