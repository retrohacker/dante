package main

import (
	"fmt"
)

func main() {
	var err error
	var inventory Inventory
	inventory, err = GetInventory()
	if err != nil {
		fmt.Printf("%v\n", err)
		return
	}
	errs := runTests(inventory)
	if len(errs) > 0 {
		fmt.Printf("# Conclusion\n\n%v tests failed.\n\n", len(errs))
		for i,err := range errs {
			fmt.Printf("%v. `%v`\n",i+1,err)
		}
	}
}
