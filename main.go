package main

import (
	"crypto/ecdsa"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	//"time"
	//"io/ioutil"
	//"github.com/ethereum/go-ethereum/ethclient"
	//"context"
)

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

type EthTransaction struct {
	Address common.Address
	Topics  []string
}

type ChargeInfo struct {
	UserID int64 `json:",string"`
	USD    float64
	Wei    int64 `json:",string"`
	AxieID int64
	Token  string
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

func connectGethOrDie() *ethclient.Client {

	url := fmt.Sprintf("wss://%s", INFURA)
	geth, geth_err := ethclient.Dial(url)
	if geth_err != nil {
		log.Fatal("Was not able to connect to the Infura Node")
	}
	return geth
}

func createUsersTable() {
	db := connectDB()

	wallets := `
	CREATE TABLE Wallets( userID integer not null primary key autoincrement, 
	address text, privateKey text, publicKey text);`

	billing := `
	CREATE TABLE Billing( userID integer not null primary key autoincrement,
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

func newUser() UserWallet {

	wallet := createWallet("")
	userID, err := saveWalletInfo(wallet)

	if err != nil {
		log.Printf("Failed to save walletinfo for a new user. Error: %s\n", err)
		return UserWallet{}
	}

	return UserWallet{
		UserID:     strconv.Itoa(userID),
		EthAddress: wallet.Address.Hex(),
	}
}

/*
	Create a wallet for the specifed use
	Save the corresponding keys to the DB.
*/
func createWallet(userID string) Wallet {
	//connect to eth provider, and create wallet return the public address.

	var ethAddress common.Address
	var privateKey *ecdsa.PrivateKey
	var pubkey *ecdsa.PublicKey

	privateKey, err := crypto.GenerateKey()

	if err != nil {
		log.Fatalf("Failed to create privtate key for user: %s. Error: %s\n", userID, err)
	}

	publicKey := privateKey.Public()
	pubkey = publicKey.(*ecdsa.PublicKey)

	ethAddress = crypto.PubkeyToAddress(*pubkey)

	wallet := Wallet{
		UserID:     userID,
		PublicKey:  *pubkey,
		PrivateKey: *privateKey,
		Address:    ethAddress,
	}
	return wallet
}

func saveWalletInfo(wallet Wallet) (int, error) {

	c := connectDB()
	var stm string
	var res sql.Result
	var err error

	if wallet.UserID != "" {
		stm = `
			INSERT INTO Wallets(userId, address, privateKey, publicKey)  
			VALUES(?,?,?,?);`

		res, err = c.Exec(
			stm,
			wallet.UserID,
			wallet.Address.Hex(),
			wallet.PrivateKeyHex(),
			wallet.PublicKeyHex(),
		)
	} else {
		stm = `INSERT INTO Wallets(address, privateKey, publicKey)
				Values(?,?,?)`

		res, err = c.Exec(
			stm,
			wallet.Address.Hex(),
			wallet.PrivateKeyHex(),
			wallet.PublicKeyHex(),
		)
	}

	if err != nil {
		log.Printf("Failed to insert new address for user: %s, ethAddress: %s\n",
			wallet.UserID, wallet.Address.Hex(),
		)
		return 0, err
	} else {
		userID, _ := res.LastInsertId()
		return int(userID), nil
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
		log.Printf("Failed to insert new billing info for user: %s, error: %s billingInfo  \n%#v\n",
			binfo.UserID, err, binfo,
		)
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
		log.Printf("Failed to query DB for wallet info of user: %s\n", userID)
		return result
	}

	defer rows.Close()
	row := rows.Next()

	if row == false {
		wallet := createWallet(userID)
		ethAddress = wallet.Address.Hex()
		_, err := saveWalletInfo(wallet)

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

func convWeiToUSD(wei int64) float64 {
	type ExchangeRates struct {
		USD float64 `json:",string"`
	}

	type Data struct {
		Currency string
		Rates    ExchangeRates
	}

	type Resp struct {
		Data Data
	}

	inEth := float64(wei) / math.Pow(10, 18)
	var coinbaseResp Resp

	resp, _ := http.Get("https://api.coinbase.com/v2/exchange-rates?currency=ETH")
	defer resp.Body.Close()

	err := json.NewDecoder(resp.Body).Decode(&coinbaseResp)
	if err != nil {
		if e, ok := err.(*json.SyntaxError); ok {
			log.Printf("syntax error at byte offset %d", e.Offset)
		}
		log.Fatalf("Failed to decode json response for eth price. Error: %s\n", err)
	}
	ethPrice := coinbaseResp.Data.Rates.USD

	//log.Printf("Price of Eth: %f. Price in wei %d. Price in Eth: %f. Price in USD: %f\n",
	//	ethPrice, wei, inEth, (ethPrice * inEth),
	//)

	return ethPrice * inEth
}

func getPrivKey(userID int64) *ecdsa.PrivateKey {
	var privKeyStr string

	c := connectDB()
	q := `SELECT privateKey FROM Wallets where userID = ?`
	rows, qerr := c.Query(q, userID)
	if qerr != nil {
		log.Printf("Failed to find private key for user: %d. Error: %s\n", userID, qerr)
	}

	defer rows.Close()
	row := rows.Next()
	if row == false {
		log.Fatalf("No rows found when querying for private key.")
	}

	err := rows.Scan(&privKeyStr)
	if err != nil {
		log.Fatalf("Failed to covert private key result into a string. Error: %s\n", err)
	}
	//log.Printf("privKeystr is: %s\n", privKeyStr)
	privKey, _ := crypto.HexToECDSA(privKeyStr)
	return privKey
}

func main() {
	//dropDB()
	//createUsersTable()
	StartServer()
}
