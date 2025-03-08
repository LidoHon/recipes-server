package helpers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
)

// PaymentRequest represents the request body for the payment API
type PaymentRequest struct {
	Amount             int    `json:"amount"`
	Currency           string `json:"currency"`
	Email              string `json:"email"`
	FirstName          string `json:"first_name"`
	PhoneNumber        string `json:"phone_number"`
	TxRef              string `json:"tx_ref"`
	CallbackURL        string `json:"callback_url"`
	ReturnURL          string `json:"return_url"`
	CustomizationTitle string `json:"customization[title]"`
	CustomizationDesc  string `json:"customization[description]"`
}

// PaymentResponse represents the structure of the payment response
type PaymentResponse struct {
	Status        bool                   `json:"status"`
	ChapaResponse map[string]interface{} `json:"chapaResponse,omitempty"`
	Message       string                 `json:"message,omitempty"`
}

// InitPayment initializes a payment and returns a response
func InitPayment(form *PaymentRequest) (*PaymentResponse, error) {

	// Prepare the request body
	body := PaymentRequest{
		Amount:             form.Amount,
		Currency:           "ETB",
		Email:              form.Email,
		FirstName:          form.FirstName,
		PhoneNumber:        form.PhoneNumber,
		TxRef:              form.TxRef,
		CallbackURL:        form.CallbackURL,
		ReturnURL:          form.ReturnURL,
		CustomizationTitle: form.CustomizationTitle,
		CustomizationDesc:  form.CustomizationDesc,
	}

	// Prepare the headers
	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", os.Getenv("CHAPA_SECRET_KEY")),
		"Content-Type":  "application/json",
	}

	// Marshal the request body to JSON
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return &PaymentResponse{Status: false, Message: "Failed to create payment request"}, err
	}

	// Create the HTTP request
	req, err := http.NewRequest("POST", os.Getenv("CHAPA_PAYMENT_ENDPOINT"), bytes.NewBuffer(jsonBody))
	if err != nil {
		return &PaymentResponse{Status: false, Message: "Failed to create request"}, err
	}
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Make the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return &PaymentResponse{Status: false, Message: err.Error()}, err
	}
	defer resp.Body.Close()

	// Read the response
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return &PaymentResponse{Status: false, Message: "Failed to read response"}, err
	}

	// Parse the response
	var chapaResponse map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &chapaResponse); err != nil {
		return &PaymentResponse{Status: false, Message: "Failed to parse response"}, err
	}
	// fmt.Println("chappa response",chapaResponse)
	// Return the response
	return &PaymentResponse{Status: true, ChapaResponse: chapaResponse}, nil

}

// VerifyPayment verifies the payment transaction using Chapa API
func VerifyPayment(txRef string) (bool, error) {
	if txRef == "" {
		return false, fmt.Errorf("txRef is empty")
	}
	// Construct the Chapa API URL with the transaction reference
	url := "https://api.chapa.co/v1/transaction/verify/" + txRef
	log.Println("URL", url)
	method := "GET"
	payload := strings.NewReader("")
	log.Println("url", url)

	// Set up HTTP client and request
	client := &http.Client{}
	req, err := http.NewRequest(method, url, payload)
	if err != nil {
		return false, fmt.Errorf("failed to create request: %v", err)
	}

	// Add authorization header with the secret key
	req.Header.Add("Authorization", "Bearer "+os.Getenv("CHAPA_SECRET_KEY"))

	// Send the request and get the response
	res, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("failed to send request: %v", err)
	}
	defer res.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return false, fmt.Errorf("failed to read response body: %v", err)
	}

	// Log the response body (for debugging purposes)
	fmt.Println(string(body))

	// Example: Check if the status in the response body is "success"
	// You need to update this part based on the actual response structure.
	var chapaResp struct {
		Status string `json:"status"`
		Data   struct {
			Status string `json:"status"`
		} `json:"data"`
	}

	// Unmarshal the JSON response into the `chapaResp` struct
	if err := json.Unmarshal(body, &chapaResp); err != nil {
		return false, fmt.Errorf("failed to parse JSON response: %v", err)
	}

	// Check the status of the payment
	if chapaResp.Data.Status == "success" {
		return true, nil
	}

	// Return error if payment verification fails
	return false, fmt.Errorf("payment verification failed: %s", chapaResp.Status)
}
