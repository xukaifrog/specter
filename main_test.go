package main

import (
	"encoding/json"
	"fmt"
	"github.com/davecgh/go-spew/spew"
	"os"
	"os/exec"
	"testing"
)

func TestMain(m *testing.M) {
	go StartServer()

	dropDB()
	createUsersTable()
	os.Exit(m.Run())
}

func TestDBCreated(t *testing.T) {
	c := connectDB()

	err := c.Ping()
	if err != nil {
		t.Fatalf("Failed to create the users database\n. Error: %v", err)
	}
}

func sendNewUserData(path string) *exec.Cmd {
	//curl -X POST -H 'Content-Type:application/json' -d @test/userData.json http://localhost:8080/newuser

	data := fmt.Sprintf("@%s", path)
	cmd := exec.Command(
		"curl",
		"-X",
		"POST",
		"-H",
		"Content-Type:application/json",
		"-d",
		data,
		"http://localhost:8080/newuser",
	)

	cmd.Dir = "/home/tyler/go/specter"
	return cmd

}

func TestNewUserPost(t *testing.T) {
	cmd := sendNewUserData("test/userData.json")
	//t.Logf("Command is: %s\n", cmd.String())
	out, err := cmd.Output()

	if err != nil {
		t.Fatalf("Failed to get the output from curl to execute test post, error: %v\n", err)
	}

	var wallet UserWallet
	perr := json.Unmarshal(out, &wallet)

	if perr != nil {
		t.Fatalf("Failed to submit userinfo. Curl Output: %s\n", out)
	}

}

func TestGetBillingInfo(t *testing.T) {
	/*
		This test will ensure that the data posted to the DB is the same as what
		has been retreieved from the db.
	*/
	cmd := *sendNewUserData("test/userData2.json")
	cmd.Run()

	//expected new user data
	var expected = BillingInfo{
		UserID:      "10124",
		BillingName: "Tyler A. Beverley",
		BillingAddress: BillingAddress{
			Address: "1152 Symphony Lane",
			City:    "Chaska",
			State:   "Minnesota",
			Zip:     "55318",
			Country: "USA",
		},
		CCInfo: CreditCardInfo{
			CCNumber:  "111111111111111",
			CVC:       "123",
			ExpirDate: "07/10",
		},
	}

	res := getBillingInfo("10124")
	if expected != res {
		t.Errorf(
			`Failed to get the correct user billing data from the database...
			Expected: %s, 
			Actual:  %s,`,
			spew.Sdump(expected),
			spew.Sdump(res),
		)
	}
}

func TestGetWalletInfo(t *testing.T) {
	/*
		This ensures that we create a wallet for a user
		when new billing details are added, and that we are able to retreive it.
	*/

	var info UserWallet

	cmd := *sendNewUserData("test/userData3.json")
	cmd.Run()

	info = getUserWallet("10125")

	if info.EthAddress == "" {
		t.Errorf("Failed to get a wallet after submiting a new user.\n Info Recieved = %#v\n", info)
	}

}

func TestUserWalletEndpoint(t *testing.T) {

	//curl http://localhost:8080/wallet/10126
	cmd := *sendNewUserData("test/userData4.json")
	cmd.Run()

	out, err := exec.Command(
		"curl",
		"http://localhost:8080/wallet/10126",
	).Output()

	if err != nil {
		t.Fatalf("Failed to get the output from curl to execute test post, error: %v\n", err)
	}

	var wallet UserWallet
	perr := json.Unmarshal(out, &wallet)

	if perr != nil {
		t.Fatalf("Failed to get userinfo. Curl Output: %s\n", out)
	}
}
