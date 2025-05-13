package handler

import (
	"net/http"

	"github.com/farmako/coupon-system/internal/domain"
	"github.com/farmako/coupon-system/internal/service"
	"github.com/gin-gonic/gin"
)

type CouponHandler struct {
	service service.CouponService
}

func NewCouponHandler(service service.CouponService) *CouponHandler {
	return &CouponHandler{service: service}
}

// GetApplicableCoupons godoc
// @Summary Get applicable coupons
// @Description Get all coupons that are applicable for the given order
// @Tags coupons
// @Accept json
// @Produce json
// @Param request body domain.CouponRequest true "Coupon Request"
// @Success 200 {array} domain.Coupon
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /coupons/applicable [get]
func (h *CouponHandler) GetApplicableCoupons(c *gin.Context) {
	var request domain.CouponRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	coupons, err := h.service.GetApplicableCoupons(c.Request.Context(), request)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, coupons)
}

// ValidateCoupon godoc
// @Summary Validate a coupon
// @Description Validate if a coupon can be applied to the given order
// @Tags coupons
// @Accept json
// @Produce json
// @Param request body domain.CouponValidationRequest true "Coupon Validation Request"
// @Success 200 {object} domain.CouponValidationResponse
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /coupons/validate [post]
func (h *CouponHandler) ValidateCoupon(c *gin.Context) {
	var request domain.CouponValidationRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.service.ValidateCoupon(c.Request.Context(), request)
	if err != nil {
		switch err.(type) {
		case *domain.CouponNotFoundError:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, response)
}
