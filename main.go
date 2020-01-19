package main

import (
	//"encoding/json"
	"database/sql"
	"fmt"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	//"io/ioutil"
	"log"
	"net/http"
	//"strconv"
	"os"
)

type BillingAddress struct {
	Address string `form:"Address"`
	City    string `form:"City"`
	State   string `form:"State"`
	Zip     string `form:"Zip"`
	Country string `form:"Country"`
}

type CreditCardInfo struct {
	CreditCardNumber     string `form:"CCNumber"`
	CardVerificationCode string `form:"CVC"`
	ExpirationDate       string `form:"ExpirDate"`
}

type BillingInfo struct {
	UserID         string         `form:"UserID"`
	BillingName    string         `form:"BillingName"`
	BillingAddress BillingAddress `form:"BillingAddress"`
	CCInfo         CreditCardInfo `form:"CreditCardInfo"`
}

type UserWallet struct {
	UserID     string
	EthAddress string
}

func ConnectDB() *(sql.DB) {
	db, err := sql.Open("sqlite3", "./users.db")
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func CreateUsersTable() {
	db := ConnectDB()

	wallets := `
	CREATE TABLE Wallets( userID integer not null primary key, 
	address text, privateKey text, seedPhrase text);`

	billing := `
	CREATE TABLE Billing( userID integer not null primary key,
	billingName text, address text, city text, state text, zip integer, country text, 
	CCNum text, CVC integer, expirDate text );`

	_, err1 := db.Exec(wallets)
	_, err2 := db.Exec(billing)

	if err1 != nil {
		log.Fatal("Failed to create wallets table", err1)
	}
	if err2 != nil {
		log.Fatal("Failed to create billing table.", err2)
	}

}

func DropDB() {
	os.Remove("./users.db")
}

/**
 *	Create a wallet for the specifed use
 *	Save the corresponding keys to the DB.
 *
 */
func CreateWallet(userID string) string {
	//connect to eth provider, and create wallet return the public address.
	var ethAddress string = "abc"
	var privateKey string = "abc"
	var seedPhrase string = "abc"

	c := ConnectDB()

	stm := `
		INSERT INTO Wallets(userId, address, privateKey, seedPhrase)  
		VALUES(?,?,?,?);`

	_, err := c.Exec(stm, userID, ethAddress, privateKey, seedPhrase)
	if err != nil {
		log.Printf("Failed to insert new address for user: %s, ethAddress: %s\n", userID, ethAddress)
	}
	return ethAddress
}

func SaveBillingInfo(binfo BillingInfo) error {

	c := ConnectDB()
	stmt := `
		INSERT INTO Billing(userID, billingName, address, 
		city, state, zip, country, CCNum, CVC, expirDate) 
		VALUES( ?, ?, ?, ?, ?, ?, ?, ?, ?, ? );`

	_, err := c.Exec(
		stmt,
		binfo.UserID,
		binfo.BillingName,
		binfo.BillingAddress.Address,
		binfo.BillingAddress.City,
		binfo.BillingAddress.State,
		binfo.BillingAddress.Zip,
		binfo.BillingAddress.Country,
		binfo.CCInfo.CreditCardNumber,
		binfo.CCInfo.CardVerificationCode,
		binfo.CCInfo.ExpirationDate,
	)

	if err != nil {
		log.Printf("Failed to insert new billing info for user: %s, billingInfo  %#v\n", binfo.UserID, binfo)
		return err
	} else {
		return nil
	}

}

func GetBillingInfo(userID string) BillingInfo {

	var binfo BillingInfo
	c := ConnectDB()

	q := `SELECT name, address, city, state, zip, country, CCnum, CVC, expirDate FROM Billing WHERE userID = ?`

	rows, err := c.Query(q, userID)
	if err != nil {
		log.Printf("Failed to query DB for %s's billing information. \n", userID)
		return binfo
	}

	defer rows.Close()
	row := rows.Next()

	if row == false {
		//if there is no billing information
		log.Printf("No Billing info found for user: %s\n", userID)
		return binfo
	}
	perr := rows.Scan(
		&binfo.BillingName,
		&binfo.BillingAddress.Address,
		&binfo.BillingAddress.City,
		&binfo.BillingAddress.State,
		&binfo.BillingAddress.Zip,
		&binfo.BillingAddress.Country,
		&binfo.CCInfo.CreditCardNumber,
		&binfo.CCInfo.CardVerificationCode,
		&binfo.CCInfo.ExpirationDate,
	)

	if perr != nil {
		log.Printf("Failed to get the billing info for user: %s\n.", userID)
	}

	return binfo
}

func GetUserWallet(userID string) UserWallet {
	//Gets the user's address
	//check databse, if not present create new

	var ethAddress string
	var result UserWallet
	c := ConnectDB()
	q := `SELECT address FROM Wallets where userID = ?;`

	rows, err := c.Query(q, userID)
	if err != nil {
		log.Printf("Failed to find wallet info for user: %s\n", userID)
		return result
	}

	defer rows.Close()
	row := rows.Next()

	if row == false {
		ethAddress = CreateWallet(userID)
	} else {
		perr := rows.Scan(&ethAddress)
		if perr != nil {
			return result
		}
	}

	result = UserWallet{
		UserID:     userID,
		EthAddress: ethAddress,
	}
	return result
}

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

func NewUser(c *gin.Context) {

	var binfo BillingInfo
	var uinfo UserWallet

	if c.ContentType() != "application/json" {
		log.Printf("Recieved a /newUser/ request from %s, but wrong content-type header.", c.ClientIP())
		c.String(400, "Expecting 'application/json' content-type header")
	}

	err := c.ShouldBind(&binfo)
	if err == nil {

		log.Printf("binfo: %#v\n", binfo)
		uinfo = GetUserWallet(binfo.UserID)
		err := SaveBillingInfo(binfo)

		if err != nil {
			log.Printf("Failed to save new billing info: %#v\n", binfo)
			c.String(500, "Error in saving user billing information")
		} else {
			c.JSON(200, uinfo)
		}
	} else {
		log.Printf("Failed to create a new user. Error: %v\n", err)
		c.String(400, "Error in parsing user infomation.")
	}
}

func StartServer() {
	r := gin.Default()

	//allow all origins
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	r.Use(cors.New(config))

	r.GET("/", Index)
	r.POST("/newuser", NewUser)

	r.Run()
}

func main() {
	DropDB()
	CreateUsersTable()
	StartServer()
}
