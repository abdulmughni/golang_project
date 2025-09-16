package stripe

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go/v81"
	"github.com/stripe/stripe-go/v81/checkout/session"
)

func CreateStripeCheckoutSession(c *gin.Context) {

	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	domain := "https://app.solutionpilot.ai"
	customerID := c.Query("stripe_customer_id")
	priceID := c.Query("price_id")

	if customerID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "stripe_customer_id is required"})
		return
	}

	if priceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "price_id is required"})
		return
	}

	params := &stripe.CheckoutSessionParams{
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			&stripe.CheckoutSessionLineItemParams{
				Price:    stripe.String(priceID),
				Quantity: stripe.Int64(1),
			},
		},
		Mode:         stripe.String(string(stripe.CheckoutSessionModeSubscription)),
		Customer:     stripe.String(customerID),
		SuccessURL:   stripe.String(domain + "?success=true"),
		CancelURL:    stripe.String(domain + "?canceled=true"),
		AutomaticTax: &stripe.CheckoutSessionAutomaticTaxParams{Enabled: stripe.Bool(false)},
	}

	s, err := session.New(params)

	if err != nil {
		log.Printf("session.New: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create session"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"url":       s.URL,
		"sessionId": s.ID,
	})
}
