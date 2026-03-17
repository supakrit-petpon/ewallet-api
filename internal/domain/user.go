package domain

type User struct{
	ID uint `json:"id"`
	Email string `json:"email" gorm:"unique; not null" validate:"required,email"`
	Password string `json:"password" gorm:"not null" validate:"required,min=8"`

	Wallet   *Wallet `json:"wallet,omitempty" gorm:"foreignKey:UserID"`
}

type UserRepository interface{
	CreateUser(user User) (uint, error)
	Transaction_CreateUser_CreateWallet(fn func(txUser UserRepository, txWallet WalletRepository) error) error
	Find(email string) (*User, error)
}