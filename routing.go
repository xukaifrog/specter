package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/stripe/stripe-go"
	"github.com/stripe/stripe-go/charge"
	"log"
	"net/http"
)

/*
 * API routing functions.
 *
 *
 */

func Ping(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}

func Index(c *gin.Context) {

	message := fmt.Sprintf(`Specter is an a payment processor that enables USD payments for crypto enabled technologies. 
	"There is a specter haunting the modern world..." - Timothy C. May.`)

	w := c.Writer
	fmt.Fprintf(w, message)
}

func SaveBilling(c *gin.Context) {

	var binfo BillingInfo
	var uinfo UserWallet

	if c.ContentType() != "application/json" {
		log.Printf("Recieved a /saveBilling/ request from %s, but wrong content-type header.", c.ClientIP())
		c.String(400, "Expecting 'application/json' content-type header")
	}

	err := c.ShouldBind(&binfo)
	if err == nil {

		//log.Printf("binfo: %#v\n", binfo)
		uinfo = getUserWallet(binfo.UserID)
		err := saveBillingInfo(binfo)

		if err != nil {
			log.Printf("Failed to save new billing info. Error: %s\n", err)
			c.String(500, "Error in saving user billing information\n")
		} else {
			c.JSON(200, uinfo)
		}
	} else {
		log.Printf("Failed to create a new user. Error: %s\n", err)
		c.String(400, "Error in parsing user infomation\n.")
	}
}

func GetUserWallet(c *gin.Context) {

	var uinfo UserWallet
	userID := c.Params.ByName("userID")
	uinfo = getUserWallet(userID)

	c.JSON(200, uinfo)
}

func Charge(c *gin.Context) {

	stripe.Key = "sk_test_2ZHt5uz8Y7dt6hEcduG59lek00ZcNJQf0X"
	var chargeInfo ChargeInfo

	if c.ContentType() != "application/json" {
		log.Printf("Recieved a charge request from %s, but wrong content-type header.",
			c.ClientIP(),
		)
	}

	err := c.ShouldBind(&chargeInfo)
	if err != nil {
		if e, ok := err.(*json.SyntaxError); ok {
			log.Printf("syntax error at byte offset %d", e.Offset)
		}
		log.Fatalf("Failed to parse response into a "+
			"chargeInfo struct. ChargeInfo: %#v.\n Error: %s\n",
			chargeInfo, err)
	}

	params := &stripe.ChargeParams{
		Amount:      stripe.Int64(chargeInfo.Ammount),
		Currency:    stripe.String(string(stripe.CurrencyUSD)),
		Description: stripe.String("Example charge"),
	}
	params.SetSource(chargeInfo.Token)
	_, stripe_err := charge.New(params)

	if stripe_err != nil {

		log.Fatalf("Failed to create charge object. Error: %s\n", err)
		c.String(400, "Failed to charge user")
	}

	buyAxie(chargeInfo)
	c.String(200, "Successfully completed order")
}

func GetAxie(c *gin.Context) {

	axie := getRandomAxie()
	c.JSON(200, axie)
}

func NewUser(c *gin.Context) {
	c.JSON(200, newUser())
}

func StartServer() {
	r := gin.Default()

	//allow all origins
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	r.Use(cors.New(config))

	r.GET("/", Index)
	r.GET("/axie", GetAxie)
	r.GET("/wallet/:userid", GetUserWallet)
	r.GET("/newuser", NewUser)
	r.POST("/savebilling", SaveBilling)
	r.POST("/charge", Charge)
	//r.POST("/buyAxie", BuyAxie)

	r.Run()
}
