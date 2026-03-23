package domain

type User struct{
	ID uint `json:"id"`
	Email string `json:"email" gorm:"unique; not null" validate:"required,email"`
	Password string `json:"password" gorm:"not null" validate:"required,min=8"`

	Wallet   *Wallet `json:"wallet,omitempty" gorm:"foreignKey:UserID"`
}

type UserRepository interface{
	Create(user User) (uint, error)
	Find(email string) (*User, error)

	//DB transaction to create user & create wallet
	ExecuteTransaction(fn func(txUser UserRepository, txWallet WalletRepository) error) error
}