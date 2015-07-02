package main

import (
	"fmt"
)

func main() {
	var err error
	var inventory Inventory
	inventory, err = GetInventory()
	if err != nil {
		fmt.Printf("%v", err)
		return
	}
	//fmt.Printf("%v \n",inventory)
	errs := runTests(inventory)
	if len(errs) > 0 {
		fmt.Printf("%v tests failed.\n", len(errs))
		fmt.Printf("%v\n", errs)
	}
	//fmt.Printf("%v",inventory)
}
