package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
)

/*
getInventory() simply reads and returns the contents of the inventory.yml file
located in the process's current working directory.
*/
func getInventory() (file []byte, err error) {
	// Begin declaring local variables
	var cwd, filename string
	// End declaring local variables

	// The next few lines attempts to get the current working directory of the
	// process and create an absolute path to the inventory.yml file located
	// there
	cwd, err = os.Getwd()
	if err != nil {
		return
	}
	filename = filepath.Join(cwd, "inventory.yml")

	// Attempt to read the file into memory
	file, err = ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return
}

/*
Inventory is used to unmarshal the inventory.yml file

NOTE: This places assumptions on the structure of the inventory.yml file. If
we want more intuitive error messages moving forward, we may want to add a
intermediary map[interface{}]interface{} object which the yaml library can
unmarshal into without errors, then have verifyInventory() handle converting
the unstructure map to the Inventory structure while verifying the content and
structure of the file.
*/
type Inventory map[string][]map[string]interface{}

/*
parseInventory takes in a raw byte array representing an inventory.yml file
unmarshals it into a go map.
*/
func parseInventory(file []byte) (obj Inventory, err error) {
	// Create a new inventory structure that will hold our object
	obj = Inventory{}

	// Unmarshal and return our new Inventory object
	err = yaml.Unmarshal(file, &obj)
	if err != nil {
		return nil, err
	}
	return
}

/*
verifyInventory ensures the structure and contents of an inventory.yml file are
correct before the application attempts to process it.
*/
func verifyInventory(inventory Inventory) (err error) {
	// Currently we do absolutely no verification, we simply allow the application
	// to fail. This function will be useful moving forward when the need for
	// more intuitive error messages arises.
	return nil
}

/*
containsDockerfile ensures that a dockerfile exists in a directory. This is
currently not used, but will serve a purpose in the verifyInventory function
moving forward.
*/
func containsDockerfile(dockerdir string) (err error) {
	var dockerDir, dockerfile string
	var file *os.File
	dockerDir, err = filepath.Abs(dockerdir)
	if err != nil {
		return
	}
	dockerfile = filepath.Join(dockerDir, "Dockerfile")
	file, err = os.Open(dockerfile)
	if err != nil {
		return
	}
	err = file.Close()
	if err != nil {
		return
	}
	return
}

/*
GetInventory is the method you should be calling when interacting with the
contents of this file. It loads in an inventory.yml file, converts it to a go
object, verifies its structure, and returns the object.
*/
func GetInventory() (inventory Inventory, err error) {
	// Begin declaring local variables
	var file []byte
	// End declaring local variables

	// Load the inventory file from disk
	file, err = getInventory()
	if err != nil {
		return nil, err
	}

	// Convert the inventory file to a go object
	inventory, err = parseInventory(file)
	if err != nil {
		return nil, err
	}

	// Verify the structure of the inventory object
	err = verifyInventory(inventory)
	if err != nil {
		return nil, err
	}
	return
}
