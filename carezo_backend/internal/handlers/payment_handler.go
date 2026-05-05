package handlers

import (
	"crypto/hmac"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"

	"github.com/delaquash/carezo/configs"
	"github.com/delaquash/carezo/internal/services"
	response "github.com/delaquash/carezo/pkg"
	"github.com/gin-gonic/gin"
)

type PaymentHandler struct {
	paymentService *services.PaymentService
	cfg            *configs.Config
}

func NewPaymentHandler(cfg *configs.Config) *PaymentHandler {
	return &PaymentHandler{
		paymentService: services.NewPaymentService(cfg.PaystackSecretKey),
		cfg:            cfg,
	}
}

// POST /api/payments/initialize
func (h *PaymentHandler) InitializePayment(c *gin.Context) {
	_, exist := c.Get("user_id")
	if !exist {
		c.JSON(401, gin.H{"error": "Unauthorized"})
		return
	}

	userEmail, _ := c.Get("user_email")

	var req struct {
		BookingID string `json:"booking_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, http.StatusBadRequest, "invalid request: "+err.Error())
		return
	}

	result, err := h.paymentService.InitializePayment(req.BookingID, userEmail.(string))

	if err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}
	response.Success(c, http.StatusOK, "payment initialize", result)
}

// GET /api/payments/verify/:reference
func (h *PaymentHandler) VerifyPayment(c *gin.Context) {
	_, exists := c.Get("user_id")

	if !exists {
		response.Error(c, http.StatusBadRequest, "unauthorized")
		return
	}

	reference := c.Param("reference")

	if reference == "" {
		response.Error(c, http.StatusBadRequest, "payment reference is required")
		return
	}

	if err := h.paymentService.VerifyPayment(reference); err != nil {
		response.Error(c, http.StatusBadRequest, err.Error())
		return
	}

	response.Success(c, http.StatusOK, "payment verified successfully", nil)
}

// POST /api/payment/webhook
func (h *PaymentHandler) HandleWebhook(c *gin.Context) {
	// read thebaw body, we need the ray bytes Paystack sent to verfy the signatures.
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read requst "})
	}

	// verify the request is from paystack using the signature
	// paystack sign the body with HMAC SHA512 using the webhook secret
	signature := c.GetHeader("X-Paystack-Signature")

	if !h.verifyPaystackSignature(body, signature) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid signature"})
		return
	}
	// parse the event type and reference from the payload
	var event struct {
		Event string `json:"event"`
		Data  struct {
			Reference string `json:"reference"`
			Status    string `json:"status"`
		} `json:"data"`
	}

	if err := json.Unmarshal(body, &event); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid JSON payload"})
		return
	}

	// handle the event

	switch event.Event {
	case "charge.success":
		// verify and update booking to confirmed after complete payment
		if err := h.paymentService.VerifyPayment(event.Data.Reference); err != nil {
			c.JSON(http.StatusOK, gin.H{"message": "webhook received, processing error: " + err.Error()})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "payment processed successfully"})

	case "charge.failed":
		// payment failed on paystack side
		// notify user and update booking to payment failed
		c.JSON(http.StatusOK, gin.H{"message": "charge failed event acknowledged, you may want to update the booking status to payment failed"})


	default:
		c.JSON(http.StatusOK, gin.H{"message": "event acknowledged"})
	}
}


func (h *PaymentHandler) verifyPaystackSignature(body []byte, signature string) bool {
	mac := hmac.New(sha512.New, []byte(h.cfg.PaystackWebhookSecret))
	mac.Write(body)
	expected := hex.EncodeToString(mac.Sum(nil))
	return hmac.Equal([]byte(expected), []byte(signature))
}