package main

import (
	zxrpc "github.com/0xProject/0x-mesh/rpc"
	"github.com/0xProject/0x-mesh/zeroex"
	"log"
)

func get0xOrders() {
	/*
		None of this function (read: shit) is needed.
		But is perserved because of sunk cost... you know
	*/
	c, dial := zxrpc.NewClient("ws://localhost:60557")

	if dial != nil {
		log.Println("Failed")
		log.Fatal(dial)
	}

	log.Println("connected")
	res, err := c.GetOrders(1, 100, "")
	if err != nil {
		log.Fatal("Failed to get the 0x orders from the mesh node.\n", err)
	}
	orderInfos := res.OrdersInfos
	decoder := zeroex.NewAssetDataDecoder()

	for _, orderInfo := range orderInfos {
		order := orderInfo.SignedOrder.Order
		//log.Printf("%#v\n", order)

		ammount := orderInfo.FillableTakerAssetAmount

		makerAssetData := order.MakerAssetData
		takerAssetData := order.TakerAssetData

		var makerAsset zeroex.ERC721AssetData
		var takerAsset zeroex.ERC20AssetData

		errMaker := decoder.Decode(makerAssetData, makerAsset)
		_ = decoder.Decode(takerAssetData, takerAsset)

		if errMaker != nil {
			log.Printf("Failed to decode asset data into an erc721. errorMaker: %s\n", errMaker)
			continue
		}
		log.Printf("maker asset: %#v, taker asset: %#v. Ammount: %d\n", makerAsset, takerAsset, ammount)
	}
	log.Printf("Finished\n")
}
