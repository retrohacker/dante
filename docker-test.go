package main

import (
	"fmt"
)

func main() {
	// Begin declaring local variables
	var err error
	var inventory Inventory
	var errs []error
	// End declaring local variables

	// Load the yml definition of images and tests
	inventory, err = GetInventory()
	if err != nil {
		// If we can't find the inventory file, there is nothing left for us to do.
		fmt.Printf("%v\n", err)
		return
	}

	// Build the images and run the tests defined in the inventory file
	errs = runTests(inventory)

	// Determine if the tests passed or failed
	if len(errs) > 0 {
		// Not all tests passed, this makes docker-test a sad panda
		fmt.Printf("# Conclusion\n\n%v tests failed.\n\n", len(errs))
		for i, err := range errs {
			fmt.Printf("%v. `%v`\n", i+1, err)
		}
	} else {
		// All tests and builds completed succesfully!
		fmt.Printf("# Conclusion\n\nall tests passed.\n\n", len(errs))
	}
}
