package service

import (
	"context"
	"sync"
	"time"

	"github.com/farmako/coupon-system/internal/domain"
	"github.com/farmako/coupon-system/internal/repository"
	"github.com/redis/go-redis/v9"
)

type CouponService interface {
	GetApplicableCoupons(ctx context.Context, req domain.CouponRequest) ([]domain.Coupon, error)
	ValidateCoupon(ctx context.Context, req domain.CouponValidationRequest) (*domain.CouponValidationResponse, error)
}

type couponService struct {
	repo  repository.CouponRepository
	redis *redis.Client
	mu    sync.Mutex
}

func NewCouponService(repo repository.CouponRepository, redis *redis.Client) CouponService {
	return &couponService{
		repo:  repo,
		redis: redis,
	}
}

func (s *couponService) GetApplicableCoupons(ctx context.Context, req domain.CouponRequest) ([]domain.Coupon, error) {
	// Get all active coupons
	coupons, err := s.repo.FindAll(ctx)
	if err != nil {
		return nil, err
	}

	var applicableCoupons []domain.Coupon
	for _, coupon := range coupons {
		if time.Now().After(coupon.ExpiryDate) {
			continue
		}

		if req.OrderValue < coupon.MinOrderValue {
			continue
		}

		if len(coupon.ApplicableMedicineIDs) > 0 {
			valid := false
			for _, medID := range req.MedicineIDs {
				for _, applicableID := range coupon.ApplicableMedicineIDs {
					if medID == applicableID {
						valid = true
						break
					}
				}
				if valid {
					break
				}
			}
			if !valid {
				continue
			}
		}

		if len(coupon.ApplicableCategories) > 0 {
			valid := false
			for _, cat := range req.Categories {
				for _, applicableCat := range coupon.ApplicableCategories {
					if cat == applicableCat {
						valid = true
						break
					}
				}
				if valid {
					break
				}
			}
			if !valid {
				continue
			}
		}

		applicableCoupons = append(applicableCoupons, coupon)
	}

	return applicableCoupons, nil
}

func (s *couponService) ValidateCoupon(ctx context.Context, req domain.CouponValidationRequest) (*domain.CouponValidationResponse, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	coupon, err := s.repo.FindByCode(ctx, req.Code)
	if err != nil {
		return nil, err
	}

	if time.Now().After(coupon.ExpiryDate) {
		return &domain.CouponValidationResponse{
			IsValid: false,
			Message: "Coupon has expired",
		}, nil
	}

	if req.OrderValue < coupon.MinOrderValue {
		return &domain.CouponValidationResponse{
			IsValid: false,
			Message: "Order value is below minimum required",
		}, nil
	}

	if len(coupon.ApplicableMedicineIDs) > 0 {
		valid := false
		for _, medID := range req.MedicineIDs {
			for _, applicableID := range coupon.ApplicableMedicineIDs {
				if medID == applicableID {
					valid = true
					break
				}
			}
			if valid {
				break
			}
		}
		if !valid {
			return &domain.CouponValidationResponse{
				IsValid: false,
				Message: "No applicable medicines in cart",
			}, nil
		}
	}

	if len(coupon.ApplicableCategories) > 0 {
		valid := false
		for _, cat := range req.Categories {
			for _, applicableCat := range coupon.ApplicableCategories {
				if cat == applicableCat {
					valid = true
					break
				}
			}
			if valid {
				break
			}
		}
		if !valid {
			return &domain.CouponValidationResponse{
				IsValid: false,
				Message: "No applicable categories in cart",
			}, nil
		}
	}

	var discount float64
	if coupon.DiscountType == "percentage" {
		discount = req.OrderValue * (coupon.DiscountValue / 100)
	} else {
		discount = coupon.DiscountValue
	}

	return &domain.CouponValidationResponse{
		IsValid:     true,
		Message:     "Coupon is valid",
		Discount:    discount,
		FinalAmount: req.OrderValue - discount,
	}, nil
}
