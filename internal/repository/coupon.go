package repository

import (
	"context"

	"github.com/farmako/coupon-system/internal/domain"
	"gorm.io/gorm"
)

type CouponRepository interface {
	FindByCode(ctx context.Context, code string) (*domain.Coupon, error)
	FindAll(ctx context.Context) ([]domain.Coupon, error)
	Create(ctx context.Context, coupon *domain.Coupon) error
	Update(ctx context.Context, coupon *domain.Coupon) error
	Delete(ctx context.Context, code string) error
}

type couponRepository struct {
	db *gorm.DB
}

func NewCouponRepository(db *gorm.DB) CouponRepository {
	return &couponRepository{db: db}
}

func (r *couponRepository) FindByCode(ctx context.Context, code string) (*domain.Coupon, error) {
	var coupon domain.Coupon
	if err := r.db.Where("code = ?", code).First(&coupon).Error; err != nil {
		return nil, err
	}
	return &coupon, nil
}

func (r *couponRepository) FindAll(ctx context.Context) ([]domain.Coupon, error) {
	var coupons []domain.Coupon
	if err := r.db.Find(&coupons).Error; err != nil {
		return nil, err
	}
	return coupons, nil
}

func (r *couponRepository) Create(ctx context.Context, coupon *domain.Coupon) error {
	return r.db.Create(coupon).Error
}

func (r *couponRepository) Update(ctx context.Context, coupon *domain.Coupon) error {
	return r.db.Save(coupon).Error
}

func (r *couponRepository) Delete(ctx context.Context, code string) error {
	return r.db.Where("code = ?", code).Delete(&domain.Coupon{}).Error
}
