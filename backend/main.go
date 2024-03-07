package main

import "github.com/opg-sirius-finance-hub/shared"

func main() {
	x := shared.NewDate("09/02/1986")
	println("your birthday is " + x.String())
}
