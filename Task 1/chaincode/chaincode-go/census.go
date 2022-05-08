package main

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/golang/protobuf/ptypes"
	"github.com/hyperledger/fabric-contract-api-go/contractapi"
)

type SmartContract struct {
	contractapi.Contract
}

type Person struct {
	FirstName string `json:"firstname"`
	LastName  string `json:"lastname"`
	City      string `json:"city"`
	Address   string `json:"address"`
	Phone     uint   `json:"phone"`
	Married   bool   `json:"married"`
}

type QueryResult struct {
	Key    string `json:"Key"`
	Record *Person
}

type Update struct {
	TxId         string    `json:"txId"`
	Timestamp    time.Time `json:"time"`
	PersonRecord *Person   `json:"record"`
}

func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	ids := [2]string{"2340001122", "1235553333"}
	people := [2]Person{
		Person{
			FirstName: "Lorem",
			LastName:  "Ipsum",
			City:      "Dolgoprudny",
			Address:   "Pervomayskaya 15",
			Phone:     89001112233,
			Married:   false},
		Person{
			FirstName: "Dolor",
			LastName:  "Sit",
			City:      "Moscow",
			Address:   "Pushkina 11",
			Phone:     89001230000,
			Married:   false},
	}

	for i, person := range people {
		personAsBytes, _ := json.Marshal(person)
		err := ctx.GetStub().PutState(ids[i], personAsBytes)

		if err != nil {
			return fmt.Errorf("failed to put to world state. %s", err.Error())
		}
	}

	return nil
}

func (s *SmartContract) CreatePerson(ctx contractapi.TransactionContextInterface, id string, firstName string, lastName string, city string, address string, phone uint, married bool) error {
	personAsBytes, _ := ctx.GetStub().GetState(id)

	if personAsBytes != nil {
		return fmt.Errorf("person with id %s already exists", id)
	}

	person := Person{
		FirstName: firstName,
		LastName:  lastName,
		City:      city,
		Address:   address,
		Phone:     phone,
		Married:   married}

	personAsBytes, err := json.Marshal(person)

	if err != nil {
		return fmt.Errorf("failed to put to world state. %s", err.Error())
	}

	return ctx.GetStub().PutState(id, personAsBytes)
}

func (s *SmartContract) QueryPerson(ctx contractapi.TransactionContextInterface, id string) (*Person, error) {
	personAsBytes, err := ctx.GetStub().GetState(id)

	if err != nil {
		return nil, fmt.Errorf("failed to read from world state. %s", err.Error())
	}

	if personAsBytes == nil {
		return nil, fmt.Errorf("person with id %s does not exist", id)
	}

	person := new(Person)
	_ = json.Unmarshal(personAsBytes, person)

	return person, nil
}

func (s *SmartContract) QueryAllPersons(ctx contractapi.TransactionContextInterface) ([]QueryResult, error) {
	startKey := ""
	endKey := ""

	resultsIterator, err := ctx.GetStub().GetStateByRange(startKey, endKey)

	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	results := []QueryResult{}

	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()

		if err != nil {
			return nil, err
		}

		person := new(Person)
		_ = json.Unmarshal(queryResponse.Value, person)

		queryResult := QueryResult{Key: queryResponse.Key, Record: person}
		results = append(results, queryResult)
	}

	return results, nil
}

func (s *SmartContract) UpdateLocation(ctx contractapi.TransactionContextInterface, id string, newCity string, newAddress string) error {
	person, err := s.QueryPerson(ctx, id)

	if err != nil {
		return err
	}

	person.City = newCity
	person.Address = newAddress

	personAsBytes, _ := json.Marshal(person)

	return ctx.GetStub().PutState(id, personAsBytes)
}

func (s *SmartContract) UpdatePhone(ctx contractapi.TransactionContextInterface, id string, newPhone uint) error {
	person, err := s.QueryPerson(ctx, id)

	if err != nil {
		return err
	}

	person.Phone = newPhone

	personAsBytes, _ := json.Marshal(person)

	return ctx.GetStub().PutState(id, personAsBytes)
}

func (s *SmartContract) UpdateMarriage(ctx contractapi.TransactionContextInterface, id string, newMarriage bool) error {
	person, err := s.QueryPerson(ctx, id)

	if err != nil {
		return err
	}

	person.Married = newMarriage

	personAsBytes, _ := json.Marshal(person)

	return ctx.GetStub().PutState(id, personAsBytes)
}

func (s *SmartContract) GetUpdatesHistory(ctx contractapi.TransactionContextInterface, id string) ([]Update, error) {
	_, err := s.QueryPerson(ctx, id)

	if err != nil {
		return nil, err
	}

	resultsIterator, err := ctx.GetStub().GetHistoryForKey(id)
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var updatesHistory []Update

	for resultsIterator.HasNext() {
		response, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		person := new(Person)
		_ = json.Unmarshal(response.Value, person)

		timestamp, _ := ptypes.Timestamp(response.Timestamp)

		update := Update{
			TxId:         response.TxId,
			Timestamp:    timestamp,
			PersonRecord: person}

		updatesHistory = append(updatesHistory, update)
	}

	return updatesHistory, nil
}

func main() {

	chaincode, err := contractapi.NewChaincode(new(SmartContract))

	if err != nil {
		fmt.Printf("Error create census chaincode: %s", err.Error())
		return
	}

	if err := chaincode.Start(); err != nil {
		fmt.Printf("Error starting census chaincode: %s", err.Error())
	}
}
