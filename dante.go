package main

import (
	"flag"
	"fmt"
	"os"
)

const version string = "1.1.0"

func main() {
	// Begin declaring local variables
	var err error
	var inventory Inventory
	var errs []error
	// End declaring local variables

	// Begin gathering runtime variables
	var threads = flag.Int("j", 1, "Number of concurrent images to build")
	var print_version = flag.Bool("v", false, "Print the version of dante")
	flag.Parse()
	// End gathering runtime variables

	if *print_version {
		fmt.Println(version)
		os.Exit(0)
	}
	// Load the yml definition of images and tests
	inventory, err = GetInventory()
	if err != nil {
		// If we can't find the inventory file, there is nothing left for us to do.
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}

	// Build the images and run the tests defined in the inventory file
	errs = runTests(*threads, inventory)

	// Determine if the tests passed or failed
	if len(errs) > 0 {
		// Not all tests passed, this makes docker-test a sad panda
		fmt.Printf("# Conclusion\n\n%v tests failed.\n\n", len(errs))
		for i, err := range errs {
			fmt.Printf("%v. `%v`\n", i+1, err)
		}
		os.Exit(1)
	} else {
		// All tests and builds completed succesfully!
		fmt.Printf("# Conclusion\n\nall tests passed.\n\n", len(errs))
		os.Exit(0)
	}
}
