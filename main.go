package main

import (
	"crypto/ecdsa"
	"database/sql"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"net/http"
	"os"
	//"encoding/json"
	//"strconv"
	//"io/ioutil"
	//"github.com/ethereum/go-ethereum/ethclient"
)

const INFURA string = "https://mainnet.infura.io/v3/55cbeaa05bf14331b55bb8416e2183f9"

type BillingAddress struct {
	Address string `form:"Address"`
	City    string `form:"City"`
	State   string `form:"State"`
	Zip     string `form:"Zip"`
	Country string `form:"Country"`
}

type CreditCardInfo struct {
	CCNumber  string `form:"CCNumber"`
	CVC       string `form:"CVC"`
	ExpirDate string `form:"ExpirDate"`
}

type BillingInfo struct {
	UserID         string         `form:"UserID"`
	BillingName    string         `form:"BillingName"`
	BillingAddress BillingAddress `form:"BillingAddress"`
	CCInfo         CreditCardInfo `form:"CCInfo"`
}

type Wallet struct {
	UserID     string
	PublicKey  ecdsa.PublicKey
	PrivateKey ecdsa.PrivateKey
	Address    common.Address
}

type UserWallet struct {
	UserID     string
	EthAddress string
}

func (w Wallet) PublicKeyHex() string {
	publicKeyBytes := crypto.FromECDSAPub(&w.PublicKey)
	pubkey := fmt.Sprint(hexutil.Encode(publicKeyBytes)[2:])
	return pubkey
}

func (w Wallet) PrivateKeyHex() string {
	privateKeyBytes := crypto.FromECDSA(&w.PrivateKey)
	privKey := fmt.Sprint(hexutil.Encode(privateKeyBytes)[2:])
	return privKey
}

func connectDB() *(sql.DB) {
	db, err := sql.Open("sqlite3", "./users.db")
	if err != nil {
		log.Fatal(err)
	}
	return db
}

func createUsersTable() {
	db := connectDB()

	wallets := `
	CREATE TABLE Wallets( userID integer not null primary key, 
	address text, privateKey text, publicKey text);`

	billing := `
	CREATE TABLE Billing( userID integer not null primary key,
	name text, address text, city text, state text, zip integer, country text, 
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

func dropDB() {
	os.Remove("./users.db")
}

/**
 *	Create a wallet for the specifed use
 *	Save the corresponding keys to the DB.
 *
 */
func createWallet(userID string) Wallet {
	//connect to eth provider, and create wallet return the public address.

	var ethAddress common.Address
	var privateKey *ecdsa.PrivateKey
	var pubkey *ecdsa.PublicKey
	//var seedPhrase string

	privateKey, err := crypto.GenerateKey()

	if err != nil {
		log.Fatalf("Failed to create privtate key for user: %s. Error: %s\n", userID, err)
	}

	publicKey := privateKey.Public()
	pubkey = publicKey.(*ecdsa.PublicKey)

	ethAddress = crypto.PubkeyToAddress(*pubkey)

	return Wallet{
		UserID:     userID,
		PublicKey:  *pubkey,
		PrivateKey: *privateKey,
		Address:    ethAddress,
	}

}

func saveWalletInfo(wallet Wallet) error {

	c := connectDB()

	stm := `
		INSERT INTO Wallets(userId, address, privateKey, publicKey)  
		VALUES(?,?,?,?);`

	_, err := c.Exec(
		stm,
		wallet.UserID,
		wallet.Address.Hex(),
		wallet.PrivateKeyHex(),
		wallet.PublicKeyHex(),
	)
	if err != nil {
		log.Printf("Failed to insert new address for user: %s, ethAddress: %s\n", wallet.UserID, wallet.Address.Hex())
		return err
	} else {
		return nil
	}
}

func saveBillingInfo(binfo BillingInfo) error {

	c := connectDB()
	stmt := `
		INSERT INTO Billing(userID, name, address, 
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
		binfo.CCInfo.CCNumber,
		binfo.CCInfo.CVC,
		binfo.CCInfo.ExpirDate,
	)

	if err != nil {
		log.Printf("Failed to insert new billing info for user: %s, error: %s billingInfo  \n%#v\n", binfo.UserID, err, binfo)
		return err
	} else {
		return nil
	}

}

func getBillingInfo(userID string) BillingInfo {

	var binfo BillingInfo
	c := connectDB()

	q := `SELECT userID, name, address, city, state, zip, country, CCnum, CVC, expirDate FROM Billing WHERE userID = ?`

	rows, err := c.Query(q, userID)
	if err != nil {
		log.Printf("Failed to query DB for %s's billing information. Error: %s \n", userID, err)
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
		&binfo.UserID,
		&binfo.BillingName,
		&binfo.BillingAddress.Address,
		&binfo.BillingAddress.City,
		&binfo.BillingAddress.State,
		&binfo.BillingAddress.Zip,
		&binfo.BillingAddress.Country,
		&binfo.CCInfo.CCNumber,
		&binfo.CCInfo.CVC,
		&binfo.CCInfo.ExpirDate,
	)

	if perr != nil {
		log.Printf("Failed to get the billing info for user: %s\n.", userID)
	}
	return binfo
}

func getUserWallet(userID string) UserWallet {
	//Gets the user's address
	//check databse, if not present create new

	var ethAddress string
	var result UserWallet
	c := connectDB()
	q := `SELECT address FROM Wallets where userID = ?;`

	rows, qerr := c.Query(q, userID)
	if qerr != nil {
		log.Printf("Failed to find wallet info for user: %s\n", userID)
		return result
	}

	defer rows.Close()
	row := rows.Next()

	if row == false {
		wallet := createWallet(userID)
		ethAddress = wallet.Address.Hex()
		err := saveWalletInfo(wallet)

		if err != nil {
			log.Printf("Failed to save wallet info to the DB. Error: %s\n", err)
		}
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

func StartServer() {
	r := gin.Default()

	//allow all origins
	config := cors.DefaultConfig()
	config.AllowAllOrigins = true
	r.Use(cors.New(config))

	r.GET("/", Index)
	r.POST("/newuser", NewUser)
	r.GET("/wallet/:userID", GetUserWallet)

	r.Run()
}

func main() {
	dropDB()
	createUsersTable()
	StartServer()
}
