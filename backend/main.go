package main

import "github.com/opg-sirius-finance-hub/api"

func main() {
	x := api.NewDate("09/02/1986")
	println("your birthday is " + x.String())
}
