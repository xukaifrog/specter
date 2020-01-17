package main

import (
	"os/exec"
	"testing"
)

func TestDBCreated(t *testing.T) {
	c := ConnectDB()

	err := c.Ping()
	if err != nil {
		t.Fatalf("Failed to create the users database\n. Error: %v", err)
	}
}

func sendNewUserData() error {
	return exec.Command(
		"curl",
		"-X POST",
		"-H 'Content-Type:application/json'",
		"-d @test/userTestData.json",
		"http://localhost:8080/newuser",
	).Run()

}

func TestNewUserPost(t *testing.T) {

	go StartServer()
	err := sendNewUserData()
	if err != nil {
		t.Errorf("Failed to create a new user using userTestData.json.\n Error: %v\n", err)
	}

}

func TestGetBillingInfo(t *testing.T) {

	//expected new user data
	var expected = BillingInfo{
		UserID:      10123,
		BillingName: "Tyler A. Beverley",
		BillingAddress: BillingAddress{
			Address: "1152 Symphony Lane",
			City:    "City",
			State:   "Minnesota",
			Zip:     55318,
			Country: "USA",
		},
		CCInfo: CreditCardInfo{
			CreditCardNumber:     "111111111111111",
			CardVerificationCode: 123,
			ExpirationDate:       "07/10",
		},
	}

	res := GetBillingInfo(10123)
	if expected != res {
		t.Errorf("Failed to get the correct user billing data from the database.")
	}

}
