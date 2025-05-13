package domain

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type UsageType string

const (
	OneTime   UsageType = "one_time"
	MultiUse  UsageType = "multi_use"
	TimeBased UsageType = "time_based"
)

type DiscountType string

const (
	Percentage DiscountType = "percentage"
	Fixed      DiscountType = "fixed"
)

// @Description Coupon information
type Coupon struct {
	ID                    uuid.UUID    `json:"id" gorm:"type:uuid;primary_key"`
	Code                  string       `json:"code" gorm:"uniqueIndex"`
	ExpiryDate            time.Time    `json:"expiry_date"`
	UsageType             UsageType    `json:"usage_type" gorm:"type:varchar(20)"`
	ApplicableMedicineIDs []string     `json:"applicable_medicine_ids" gorm:"type:text[]"`
	ApplicableCategories  []string     `json:"applicable_categories" gorm:"type:text[]"`
	MinOrderValue         float64      `json:"min_order_value"`
	ValidTimeWindow       *TimeWindow  `json:"valid_time_window,omitempty" gorm:"type:jsonb"`
	TermsAndConditions    string       `json:"terms_and_conditions"`
	DiscountType          DiscountType `json:"discount_type" gorm:"type:varchar(20)"`
	DiscountValue         float64      `json:"discount_value"`
	MaxUsagePerUser       int          `json:"max_usage_per_user"`
	CreatedAt             time.Time    `json:"created_at"`
	UpdatedAt             time.Time    `json:"updated_at"`
}

// @Description Time window for coupon validity
type TimeWindow struct {
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

// @Description Request to get applicable coupons
type CouponRequest struct {
	MedicineIDs []string `json:"medicine_ids"`
	Categories  []string `json:"categories"`
	OrderValue  float64  `json:"order_value"`
	UserID      string   `json:"user_id"`
}

// @Description Request to validate a coupon
type CouponValidationRequest struct {
	Code        string   `json:"code"`
	MedicineIDs []string `json:"medicine_ids"`
	Categories  []string `json:"categories"`
	OrderValue  float64  `json:"order_value"`
	UserID      string   `json:"user_id"`
}

// @Description Response for coupon validation
type CouponValidationResponse struct {
	IsValid     bool    `json:"is_valid"`
	Message     string  `json:"message"`
	Discount    float64 `json:"discount"`
	FinalAmount float64 `json:"final_amount"`
}

type CouponNotFoundError struct {
	Code string
}

func (e *CouponNotFoundError) Error() string {
	return fmt.Sprintf("coupon not found: %s", e.Code)
}

func (tw TimeWindow) Value() (driver.Value, error) {
	return json.Marshal(tw)
}

func (tw *TimeWindow) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("failed to unmarshal TimeWindow value: %v", value)
	}
	return json.Unmarshal(bytes, tw)
}

type CartItem struct {
	ID       string  `json:"id"`
	Category string  `json:"category"`
	Price    float64 `json:"price"`
}

type Discount struct {
	ItemsDiscount   float64 `json:"items_discount"`
	ChargesDiscount float64 `json:"charges_discount"`
}

type ApplicableCouponsRequest struct {
	CartItems  []CartItem `json:"cart_items"`
	OrderTotal float64    `json:"order_total"`
	Timestamp  time.Time  `json:"timestamp"`
}

type ApplicableCoupon struct {
	CouponCode    string  `json:"coupon_code"`
	DiscountValue float64 `json:"discount_value"`
}

type ApplicableCouponsResponse struct {
	ApplicableCoupons []ApplicableCoupon `json:"applicable_coupons"`
}
