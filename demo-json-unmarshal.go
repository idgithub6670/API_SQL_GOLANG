package main

import (
	"encoding/json"
	"fmt"
	"log"
)

type employee struct {
	ID           int
	EmployeeName string
}

func main() {

	e := employee{}

	err := json.Unmarshal([]byte(`{"ID":101,"EmployeeName":"thana"}`), &e)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(e.ID)
}
