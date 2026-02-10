package services

type PaymentService struct {
	paystackSecretKey string
	paystackBaseURL string
	bookingService *BookingService
}


func NewPaymentService(paystackSecretKey string) *PaymentService {
	return &PaymentService{
		paystackSecretKey: paystackSecretKey,
		paystackBaseURL: "https://api.paystack.co",
		bookingService: NewBookingService(),
	}
}

type PaystackInitializeResponse struct {
	Status bool `json:"status"`
	Message string `json:"message"`
	Data struct {
		AuthorizationURL string `json:"authorization_url"`
		AccessCode string `json:"access_code"`
		Reference string `json:"reference"`
	} `json:"data"`
}


type PaystackVerifyResponse struct {
	Status bool `json:"status"`
	Message string `json:"message"`
	Data struct {
		ID       int64    `json:"id"`
		Domain   string   `json:"domain"`
		Status   string   `json:"status"`
		Reference string  `json:"reference"`
		Amount    int64    `json:"amount"`
		PaidAt    string   `json:"paid_at"`
		Channel   string   `json:"channel"`
		Currency  string   `json:"currency"`
		Customer  struct {
			Email string `json:"email"`
		} `json:"customer"`
	} `json:"data"`
}


func (s *PaymentService) InitializePayment(bookingID string, userEmail string){
	
}
	
