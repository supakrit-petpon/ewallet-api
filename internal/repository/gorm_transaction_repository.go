package repository

import (
	"errors"
	"piano/e-wallet/internal/domain"

	"gorm.io/gorm"
)

type GormTransactionRepository struct{
	db *gorm.DB
}

func NewGormTransactionRepository(db *gorm.DB) domain.TransactionRepository{
	return &GormTransactionRepository{db: db}
}

func (r *GormTransactionRepository) Create(tx *domain.Transaction) error {
	result := r.db.Create(&tx)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return domain.ErrConflictTransactionRefId
		}
		if errors.Is(result.Error, gorm.ErrForeignKeyViolated) {
			return domain.ErrNotFoundWallet
		}
		
		return domain.ErrInternalServerError
	}

	return nil
}

func (r *GormTransactionRepository) Update(id uint, status string) (*domain.Transaction, error) {
	trans := new(domain.Transaction)

    // 1. ดึงข้อมูลปัจจุบันออกมาก่อน
    if err := r.db.First(trans, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound){
			return nil, domain.ErrNotFoundTransaction
		}

		return nil, domain.ErrInternalServerError
    }

    // 2. แก้ไขค่าใน Memory
    trans.Status = status

    // 3. บันทึกลง Database (GORM จะ Update เฉพาะฟิลด์ที่เปลี่ยน)
    if err := r.db.Save(trans).Error; err != nil {
        return nil, domain.ErrInternalServerError
    }

    return trans, nil
}

func (r *GormTransactionRepository) Get(refId string) (*domain.Transaction, error) {
	transaction := new(domain.Transaction)

    if err := r.db.First(transaction, "reference_id = ?", refId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound){
			return nil, domain.ErrNotFoundTransaction
		}

		return nil, domain.ErrInternalServerError
    }

    return transaction, nil
}


func (r *GormTransactionRepository) GetAll(walletId uint) ([]domain.Transaction, error) {
	var transactions []domain.Transaction
	err := r.db.Where("source_id = ? OR destination_id = ?", walletId, walletId).Order("created_at DESC").Find(&transactions).Error
	if err != nil {
		return nil, domain.ErrInternalServerError
	}
	
	return transactions, nil
}
