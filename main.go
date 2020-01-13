package main

import (
	//"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	//"io/ioutil"
	//"log"
	"net/http"
)

func Ping(c *gin.Context) {
	c.String(http.StatusOK, "pong")
}

func Index(c *gin.Context) {

	message := fmt.Sprintf(`Specter is an a payment processor that enables USD payments for crypto enabled technologies. 
	"There is a specter haunting the modern world..." - Timothy C. May.`)

	w := c.Writer
	fmt.Fprintf(w, message)
}

// Should there be an accept payment for an unknown user, and a seperate one for a known user?
func ChargeCustomer(c *gin.Context) {

	// Accept Payment takes the following arguments in a post request.
	//		Billing Address {street, house number, city, state, zip, country }
	//		Name { First, Last }
	//		CC Info { CC num, expiration date, CVC }

}

func getUserAddress(c *gin.Context) {
	//gets the ethereum address associated with this userID.

}

func main() {

	r := gin.Default()
	r.GET("/", Index)
	r.POST("/charge", ChargeCustomer)

	r.Run()
}
