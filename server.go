package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/stripe/stripe-go/v71"
	"github.com/stripe/stripe-go/v71/customer"
	"github.com/stripe/stripe-go/v71/webhook"
)

func main() {
	if err := godotenv.Load("./.env"); err != nil {
		log.Fatalf("godotenv.Load: %v", err)
	}
	e := echo.New()
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	// basic setup
	e.Use(middleware.Recover())
	e.Use(middleware.Logger())
	e.Static("/", os.Getenv("STATIC_DIR"))

	// router
	e.POST("/webhook", handleWebhook)
	e.Logger.Panic(e.Start(":4242"))
}

func handleWebhook(c echo.Context) error {
	b, err := ioutil.ReadAll(c.Request().Body)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	event, err := webhook.ConstructEvent(b, c.Request().Header.Get("Stripe-Signature"), os.Getenv("STRIPE_WEBHOOK_SECRET"))
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	if event.Type != "checkout.session.completed" {
		return c.JSON(http.StatusBadRequest, fmt.Errorf("wrong event type"))
	}
	cust, err := customer.Get(event.GetObjectValue("customer"), nil)
	if err != nil {
		return c.JSON(http.StatusBadRequest, err)
	}
	if event.GetObjectValue("display_items", "0", "custom") != "" &&
		event.GetObjectValue("display_items", "0", "custom", "name") == "Pasha e-book" {
		return c.JSON(http.StatusOK, "🔔 Customer is subscribed and bought an e-book! Send the e-book to %s"+cust.Email)
	}
	return c.JSON(http.StatusOK, "🔔 Customer is subscribed but did not buy an e-book.")
}

func handleCreateCustomer(c echo.Context) error {
	var req struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("json.NewDecoder.Decode: %v", err)
		return
	}
	params := &stripe.CustomerParams{
		Email: stripe.String(req.Email),
	}
	c, err := customer.New(params)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Printf("customer.New: %v", err)
		return
	}
	writeJSON(w, struct {
		Customer *stripe.Customer `json:"customer"`
	}{
		Customer: c,
	})
}
