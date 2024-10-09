package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/sony/gobreaker"
)

type Response struct {
	Message string `json:"message"`
	Success bool   `json:"success"`
}

// Simulate primary payment gateway
func primaryPaymentGateway() error {
	// Simulate failure for the primary gateway
	_, err := http.Get("http://localhost:6666/payment")
	if err != nil {
		time.Sleep(time.Second / 2)
		log.Println("primary payment gateway failed")
		return errors.New("primary payment gateway failed")
	}
	log.Println("primary payment gateway success")
	return nil
}

// Simulate secondary payment gateway
func secondaryPaymentGateway() error {
	log.Println("secondary payment gateway success")
	return nil
}

var cb *gobreaker.CircuitBreaker

func init() {
	// Circuit breaker settings
	cbSettings := gobreaker.Settings{
		Name:     "PaymentGatewayCircuitBreaker",
		Interval: 5 * time.Second, // Reset interval for failure count
		Timeout:  5 * time.Second, // Timeout before transitioning to half-open
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 3 // Trip after 3 consecutive failures
		},
	}

	// Create circuit breaker instance
	cb = gobreaker.NewCircuitBreaker(cbSettings)
}

// function with circuit breaker
// Function to handle the payment logic
func processPayment() (string, error) {
	// Try using the primary payment gateway with circuit breaker
	_, err := cb.Execute(func() (interface{}, error) {
		err := primaryPaymentGateway()
		if err != nil {
			return nil, err
		}
		return "Primary gateway success", nil
	})

	// If primary gateway failed, fallback to secondary gateway
	if err != nil {
		err := secondaryPaymentGateway()
		if err != nil {
			return "", errors.New("payment failed through both gateways")
		}
		return "Payment succeeded through secondary gateway", nil
	}
	return "Payment succeeded through primary gateway", nil
}

// function without circuit breaker
// func processPayment() (string, error) {
// 	err := primaryPaymentGateway()
// 	// If primary gateway failed, fallback to secondary gateway
// 	if err != nil {
// 		err := secondaryPaymentGateway()
// 		if err != nil {
// 			return "", errors.New("payment failed through both gateways")
// 		}
// 		return "Payment succeeded through secondary gateway", nil
// 	}
// 	return "Payment succeeded through primary gateway", nil
// }

func getResponse(w http.ResponseWriter, message string, statusCode int) {
	success := true
	if statusCode > 299 {
		success = false
	}
	response := Response{Message: message, Success: success}
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response)
}

// Payment API Handler
func paymentHandler(w http.ResponseWriter, r *http.Request) {
	// Process the payment using the refactored function
	result, err := processPayment()

	w.Header().Set("Content-Type", "application/json")
	if err != nil {
		getResponse(w, err.Error(), http.StatusInternalServerError)
		return
	}

	getResponse(w, result, http.StatusOK)
}

func main() {
	// Set up HTTP routes
	http.HandleFunc("/pay", paymentHandler)

	// Start the server
	log.Println("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Server failed to start:", err)
	}
}
