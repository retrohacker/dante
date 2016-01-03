package main

import (
	"fmt"
	"github.com/wblankenship/cli"
	"os"
)

const version string = "1.1.0"

// Inventory is static and global
// We will initialize it once and then use it throughout the app
var inventory Inventory

func main() {

	/* Define cli commands and flags */
	app := cli.NewApp()
	app.Commands = []cli.Command{
		{
			Name:   "test",
			Usage:  "Build images and run tests defined in inventory.yml",
			Action: test,
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "retries,r",
					Usage: "Retry on failure",
					Value: 0,
				},
				cli.IntFlag{
					Name:  "parallel,j",
					Usage: "Run parallel jobs",
					Value: 1,
				},
			},
		},
		{
			Name:   "push",
			Usage:  "Push local images to remote registry",
			Action: push,
			Flags: []cli.Flag{
				cli.IntFlag{
					Name:  "retries,r",
					Usage: "Retry on failure",
					Value: 0,
				},
				cli.IntFlag{
					Name:  "parallel,j",
					Usage: "Run parallel jobs",
					Value: 1,
				},
			},
		},
	}

	app.Version = version

	app.Run(os.Args)

}

/*
populateInventory initializes the global state of the application as directed
by the user's inventory.yaml file
*/
func populateInventory() {
	// Load the yml definition of images and tests
	var err error
	inventory, err = GetInventory()

	if err != nil {
		// If we can't find the inventory file, there is nothing left for us to do.
		fmt.Printf("%v\n", err)
		os.Exit(1)
	}
}

func test(c *cli.Context) {
	populateInventory()

	opts := TestOpts{
		Threads: c.Int("parallel"),
		Retries: c.Int("retries"),
	}

	/* Scrub Input */
	if opts.Threads < 1 {
		opts.Threads = 1
	}

	if opts.Retries < 0 {
		opts.Retries = 0
	}

	// Build the images and run the tests defined in the inventory file
	errs := runTests(inventory, opts)

	// Determine if the tests passed or failed
	if errs > 0 {
		// Not all tests passed, this makes docker-test a sad panda
		fmt.Printf("# Conclusion\n\n%v tests failed.\n\n", errs)
		os.Exit(1)
	} else {
		// All tests and builds completed succesfully!
		fmt.Printf("# Conclusion\n\nall tests passed.\n\n")
		os.Exit(0)
	}

}

func push(*cli.Context) {
	populateInventory()

	fmt.Println("# Push not implemented yet")
	os.Exit(1)
}
