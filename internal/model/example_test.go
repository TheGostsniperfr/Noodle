package model_test

import (
	"fmt"

	"github.com/TheGostsniperfr/Noodle/internal/model"
)

func ExampleCheck() {
	s, err := model.LoadSystem("testdata/broken")
	if err != nil {
		panic(err)
	}
	for _, f := range model.Check(s) {
		fmt.Println(f)
	}
	// Output:
	// layouts/request-path.yaml: c-api-db: endpoint "api.east": side "east", want left, right, top or bottom
	// layouts/request-path.yaml: c-api-db: lane "bus" is not defined
	// layouts/request-path.yaml: db: included element has no position
	// model.yaml: c-api-db: port "postgres" is not a port of db
	// model.yaml: x-db-api: enforced_by "netpol-default-deny" does not exist in the model
	// views/request-path.yaml: c-user-api: step is not a connection of the model
}
