package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/gateway"
	"github.com/thatisuday/commando"
)

func main() {
	commando.
		SetExecutableName("client").
		SetVersion("1.0.0").
		SetDescription("Interact with census chaincode")

	commando.
		Register("QueryPerson").
		AddArgument("id", "Person id", "").
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
			contract := connectContract()
			result, err := contract.EvaluateTransaction("QueryPerson", args["id"].Value)
			if err != nil {
				fmt.Printf("Failed to evaluate transaction: %s\n", err)
				os.Exit(1)
			}
			fmt.Println(string(result))
		})

	commando.
		Register("QueryAllPersons").
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
			contract := connectContract()
			result, err := contract.EvaluateTransaction("QueryAllPersons")
			if err != nil {
				fmt.Printf("Failed to evaluate transaction: %s\n", err)
				os.Exit(1)
			}
			fmt.Println(string(result))
		})

	commando.
		Register("CreatePerson").
		AddArgument("id", "Person id", "").
		AddFlag("fname, f", "Last name", commando.String, nil).
		AddFlag("lname, l", "First name", commando.String, nil).
		AddFlag("city, c", "New current city", commando.String, nil).
		AddFlag("address, a", "New current address", commando.String, nil).
		AddFlag("phone, p", "Personal phone number in format \"8XXXXXXXXXX\"", commando.String, "").
		AddFlag("married, m", "Is the person married", commando.String, "false").
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
			contract := connectContract()

			lname, _ := flags["lname"].GetString()
			fname, _ := flags["fname"].GetString()
			city, _ := flags["city"].GetString()
			address, _ := flags["address"].GetString()
			phone, _ := flags["phone"].GetString()
			married, _ := flags["married"].GetString()
			result, err := contract.SubmitTransaction(
				"CreatePerson",
				args["id"].Value,
				lname,
				fname,
				city,
				address,
				phone,
				married)
			if err != nil {
				fmt.Printf("Failed to evaluate transaction: %s\n", err)
				os.Exit(1)
			}
			fmt.Println(string(result))
		})

	commando.
		Register("UpdateLocation").
		AddArgument("id", "Person id", "").
		AddFlag("city,c", "New current city", commando.String, "").
		AddFlag("address,a", "New current address", commando.String, "").
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
			contract := connectContract()

			city, _ := flags["city"].GetString()
			address, _ := flags["address"].GetString()

			result, err := contract.SubmitTransaction(
				"UpdateLocation",
				args["id"].Value,
				city,
				address)
			if err != nil {
				fmt.Printf("Failed to evaluate transaction: %s\n", err)
				os.Exit(1)
			}
			fmt.Println(string(result))
		})

	commando.
		Register("UpdatePhone").
		AddArgument("id", "Person id", "").
		AddFlag("phone,p", "New personal phone", commando.String, nil).
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
			contract := connectContract()

			phone, _ := flags["phone"].GetString()
			result, err := contract.SubmitTransaction(
				"UpdatePhone",
				args["id"].Value,
				phone)
			if err != nil {
				fmt.Printf("Failed to evaluate transaction: %s\n", err)
				os.Exit(1)
			}
			fmt.Println(string(result))
		})

	commando.
		Register("UpdateMarriage").
		AddArgument("id", "Person id", "").
		AddFlag("married,m", "If person is married now", commando.String, nil).
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
			contract := connectContract()

			married, _ := flags["married"].GetString()
			result, err := contract.SubmitTransaction(
				"UpdateMarriage",
				args["id"].Value,
				married)
			if err != nil {
				fmt.Printf("Failed to evaluate transaction: %s\n", err)
				os.Exit(1)
			}
			fmt.Println(string(result))
		})

	commando.
		Register("GetUpdatesHistory").
		AddArgument("id", "Person id", "").
		SetAction(func(args map[string]commando.ArgValue, flags map[string]commando.FlagValue) {
			contract := connectContract()

			result, err := contract.EvaluateTransaction(
				"GetUpdatesHistory",
				args["id"].Value)

			if err != nil {
				fmt.Printf("Failed to evaluate transaction: %s\n", err)
				os.Exit(1)
			}
			fmt.Println(string(result))
		})
	// parse command-line arguments
	commando.Parse(nil)
}

func connectContract() *gateway.Contract {
	os.Setenv("DISCOVERY_AS_LOCALHOST", "true")
	wallet, err := gateway.NewFileSystemWallet("wallet")
	if err != nil {
		fmt.Printf("Failed to create wallet: %s\n", err)
		os.Exit(1)
	}

	if !wallet.Exists("client") {
		err = populateWallet(wallet)
		if err != nil {
			fmt.Printf("Failed to populate wallet contents: %s\n", err)
			os.Exit(1)
		}
	}

	ccpPath := filepath.Join(
		"..",
		"network",
		"organizations",
		"peerOrganizations",
		"org1.example.com",
		"connection-org1.yaml",
	)

	gw, err := gateway.Connect(
		gateway.WithConfig(config.FromFile(filepath.Clean(ccpPath))),
		gateway.WithIdentity(wallet, "appUser"),
	)
	if err != nil {
		fmt.Printf("Failed to connect to gateway: %s\n", err)
		os.Exit(1)
	}
	defer gw.Close()

	network, err := gw.GetNetwork("mychannel")
	if err != nil {
		fmt.Printf("Failed to get network: %s\n", err)
		os.Exit(1)
	}

	return network.GetContract("census")
}

func populateWallet(wallet *gateway.Wallet) error {
	credPath := filepath.Join(
		"..",
		"network",
		"organizations",
		"peerOrganizations",
		"org1.example.com",
		"users",
		"User1@org1.example.com",
		"msp",
	)

	certPath := filepath.Join(credPath, "signcerts", "cert.pem")
	// read the certificate pem
	cert, err := ioutil.ReadFile(filepath.Clean(certPath))
	if err != nil {
		return err
	}

	keyDir := filepath.Join(credPath, "keystore")
	// there's a single file in this dir containing the private key
	files, err := ioutil.ReadDir(keyDir)
	if err != nil {
		return err
	}
	if len(files) != 1 {
		return errors.New("keystore folder should have contain one file")
	}
	keyPath := filepath.Join(keyDir, files[0].Name())
	key, err := ioutil.ReadFile(filepath.Clean(keyPath))
	if err != nil {
		return err
	}

	identity := gateway.NewX509Identity("Org1MSP", string(cert), string(key))

	err = wallet.Put("appUser", identity)
	if err != nil {
		return err
	}
	return nil
}
