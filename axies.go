package main

import (
	marketplace "./contracts/Marketplace"
	"encoding/json"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	//"github.com/ethereum/go-ethereum/ethclient"
	"log"
	"math/big"
	"math/rand"
	"net/http"
	"strconv"
)

const MARKETPLACE = "0xF4985070Ce32b6B1994329DF787D1aCc9a2dd9e2"
const NFT = "0xf5b0a3efb8e8e4c201e2a935f110eaaf3ffecb8d"

type AxieAuction struct {
	BuyNowPrice string
	PriceInUSD  float64
}

type Axie struct {
	Id      int
	Name    string
	Image   string
	Owner   string
	Auction AxieAuction
}

func getRandomAxie() Axie {

	type Response struct {
		TotalAxies int
		Axies      []Axie
	}
	var axies Response

	res, err := http.Get("https://axieinfinity.com/api/v2/axies?sale=1")
	if err != nil {
		log.Fatalf("Failed to request a list of axies on sale. Error: %s.\n", err)
	}

	defer res.Body.Close()
	perr := json.NewDecoder(res.Body).Decode(&axies)
	if perr != nil {
		log.Printf("Failed to parse axie response. Error: %s\n", perr)
	}

	rand := rand.Intn(len(axies.Axies))

	chosen := axies.Axies[rand]
	priceInEth, _ := strconv.ParseInt(chosen.Auction.BuyNowPrice, 10, 64)

	usdPrice := convWeiToUSD(priceInEth)
	chosen.Auction.PriceInUSD = usdPrice
	//log.Printf("Price in USD is: %f\n", usdPrice)
	return chosen
}

func buyAxie(order ChargeInfo) {

	mplace := common.HexToAddress(MARKETPLACE)
	nft := common.HexToAddress(NFT)
	geth := connectGethOrDie()
	instance, err := marketplace.NewMarketplace(mplace, geth)

	if err != nil {
		log.Fatalf("Failed to create an instance of the axie marketplace contract. Error %s\n", err)
	}

	//privKey := getPrivKey(order.UserID)
	privKey := getPrivKey(1)
	keyedTransactor := bind.NewKeyedTransactor(privKey)
	keyedTransactor.GasLimit = 305000
	keyedTransactor.GasPrice = big.NewInt(8000000000)
	keyedTransactor.Value = big.NewInt(order.Wei)

	session := marketplace.MarketplaceSession{
		Contract:     instance,
		CallOpts:     bind.CallOpts{},
		TransactOpts: *keyedTransactor,
	}

	_, txerr := session.Bid(nft, big.NewInt(int64(order.AxieID)))
	if txerr != nil {
		log.Printf("TransactOpts: %#v\n", session.TransactOpts)
		log.Printf("Failed to create a bid for axie %d, error: %s\n", order.AxieID, txerr)
	} else {
		inventoryURL := fmt.Sprintf("https://axieinfinity.com/api/v2/addresses/%s/axies",
			(*keyedTransactor).From.Hex())
		log.Printf("Successfully bought axie number %d. URL: %s\n", order.AxieID, inventoryURL)
	}
}

func getFastGas() int64 {
	var data map[string]interface{}

	resp, err := http.Get("https://ethgasstation.info/json/ethgasAPI.json")
	if err != nil {
		log.Fatal("Failed to get gas station info. Error: %s\n", err)
	}
	defer resp.Body.Close()

	perror := json.NewDecoder(resp.Body).Decode(&data)
	if perror != nil {
		log.Fatal("Failed to parse gas station info into json. Error: %s\n", perror)
	}
	fast, _ := data["fast"].(int64)
	return (fast / 10)
}
