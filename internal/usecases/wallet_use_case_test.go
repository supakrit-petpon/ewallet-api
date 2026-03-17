package usecases

import (
	"piano/e-wallet/internal/domain"
)

type mockWalletRepo struct{
	createWalletFunc func(wallet domain.Wallet) error
}

func (m *mockWalletRepo) CreateWallet(wallet domain.Wallet) error{
	return m.createWalletFunc(wallet)
}