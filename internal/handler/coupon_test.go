package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/farmako/coupon-system/internal/domain"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockCouponService struct {
	mock.Mock
}

func (m *MockCouponService) GetApplicableCoupons(ctx context.Context, req domain.CouponRequest) ([]domain.Coupon, error) {
	args := m.Called(ctx, req)
	return args.Get(0).([]domain.Coupon), args.Error(1)
}

func (m *MockCouponService) ValidateCoupon(ctx context.Context, req domain.CouponValidationRequest) (*domain.CouponValidationResponse, error) {
	args := m.Called(ctx, req)
	return args.Get(0).(*domain.CouponValidationResponse), args.Error(1)
}

func TestGetApplicableCoupons(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCouponService)
	handler := NewCouponHandler(mockService)
	router := gin.New()
	router.GET("/coupons/applicable", handler.GetApplicableCoupons)

	t.Run("success", func(t *testing.T) {
		request := domain.CouponRequest{
			MedicineIDs: []string{"med1", "med2"},
			Categories:  []string{"cat1"},
			OrderValue:  100.0,
			UserID:      "user1",
		}

		expectedCoupons := []domain.Coupon{
			{
				Code:          "TEST123",
				ExpiryDate:    time.Now().Add(24 * time.Hour),
				UsageType:     "one_time",
				DiscountType:  "percentage",
				DiscountValue: 10.0,
				MinOrderValue: 50.0,
			},
		}

		mockService.On("GetApplicableCoupons", mock.Anything, request).Return(expectedCoupons, nil)

		// Make request
		body, _ := json.Marshal(request)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/coupons/applicable", bytes.NewBuffer(body))
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response []domain.Coupon
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedCoupons, response)
	})

	t.Run("invalid request", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/coupons/applicable", bytes.NewBufferString("invalid json"))
		router.ServeHTTP(w, req)

		// Assertions
		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}

func TestValidateCoupon(t *testing.T) {
	gin.SetMode(gin.TestMode)
	mockService := new(MockCouponService)
	handler := NewCouponHandler(mockService)
	router := gin.New()
	router.POST("/coupons/validate", handler.ValidateCoupon)

	// Test cases
	t.Run("success", func(t *testing.T) {
		// Mock data
		request := domain.CouponValidationRequest{
			Code:        "TEST123",
			MedicineIDs: []string{"med1", "med2"},
			Categories:  []string{"cat1"},
			OrderValue:  100.0,
			UserID:      "user1",
		}

		expectedResponse := &domain.CouponValidationResponse{
			IsValid:     true,
			Message:     "Coupon is valid",
			Discount:    10.0,
			FinalAmount: 90.0,
		}

		mockService.On("ValidateCoupon", mock.Anything, request).Return(expectedResponse, nil)

		body, _ := json.Marshal(request)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/coupons/validate", bytes.NewBuffer(body))
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
		var response domain.CouponValidationResponse
		err := json.Unmarshal(w.Body.Bytes(), &response)
		assert.NoError(t, err)
		assert.Equal(t, expectedResponse, &response)
	})

	t.Run("coupon not found", func(t *testing.T) {
		request := domain.CouponValidationRequest{
			Code:        "INVALID",
			MedicineIDs: []string{"med1"},
			Categories:  []string{"cat1"},
			OrderValue:  100.0,
			UserID:      "user1",
		}

		mockService.On("ValidateCoupon", mock.Anything, request).Return(nil, &domain.CouponNotFoundError{Code: "INVALID"})

		body, _ := json.Marshal(request)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/coupons/validate", bytes.NewBuffer(body))
		router.ServeHTTP(w, req)

		assert.Equal(t, http.StatusNotFound, w.Code)
	})
}
