package main

import (
	"json"
	"os"
	"os/exec"
	"testing"
)

func TestMain(m *testing.M) {
	go StartServer()

	DropDB()
	CreateUsersTable()
	os.Exit(m.Run())
}

func TestDBCreated(t *testing.T) {
	c := ConnectDB()

	err := c.Ping()
	if err != nil {
		t.Fatalf("Failed to create the users database\n. Error: %v", err)
	}
}

func sendNewUserData() *exec.Cmd {
	//curl -X POST -H 'Content-Type:application/json' -d @test/userData.json http://localhost:8080/newuser

	return exec.Command(
		"curl",
		"-X POST",
		"-H 'Content-Type:application/json'",
		"-d @test/userData.json",
		"http://localhost:8080/newuser",
	)

}

func TestNewUserPost(t *testing.T) {
	cmd := sendNewUserData()
	stdout, err := cmd.StdoutPipe()

	if err != nil {
		t.Fatalf("Failed to get the stdout pipe to execute post, error: %v\n", err)
	}

	var wallet UserWallet
	perr := json.NewDecoder(stdout).Decode(&wallet)

	if perr != nil {
		t.Fatalf("Failed to submit a submit userinfo. Error: %v\n", err)
	}

}

func TestGetBillingInfo(t *testing.T) {
	/*
		This test will ensure that the data posted to the DB is the same as what
		has been retreieved from the db.
	*/
	sendNewUserData()

	//expected new user data
	var expected = BillingInfo{
		UserID:      "10123",
		BillingName: "Tyler A. Beverley",
		BillingAddress: BillingAddress{
			Address: "1152 Symphony Lane",
			City:    "City",
			State:   "Minnesota",
			Zip:     "55318",
			Country: "USA",
		},
		CCInfo: CreditCardInfo{
			CreditCardNumber:     "111111111111111",
			CardVerificationCode: "123",
			ExpirationDate:       "07/10",
		},
	}

	res := GetBillingInfo("10123")
	if expected != res {
		t.Errorf(
			`Failed to get the correct user billing data from the database...
			Expected: %#v, 
			Actual:   %#v,`,
			expected,
			res,
		)
	}
}
