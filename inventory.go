package main

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path/filepath"
	//"fmt"
)

func getInventory() (file []byte, err error) {
	var cwd, filename string
	cwd, err = os.Getwd()
	if err != nil {
		return
	}
	filename = filepath.Join(cwd, "inventory.yml")
	file, err = ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	return
}

type Inventory map[string][]map[string]interface{}

func parseInventory(file []byte) (obj Inventory, err error) {
	obj = Inventory{}
	err = yaml.Unmarshal(file, &obj)
	if err != nil {
		return nil, err
	}
	return
}

func verifyInventory(inventory Inventory) (err error) {
	/*
		for _,image := range inventory["images"] {
			// If path is present, ensure that it leads to a valid Dockerfile
			if image["path"]!=nil {
				err = containsDockerfile(image["path"].(string))
				if err != nil {
					return
				}
			}
		}
	*/
	return nil
}

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

func GetInventory() (inventory Inventory, err error) {
	var file []byte

	file, err = getInventory()
	if err != nil {
		return nil, err
	}

	inventory, err = parseInventory(file)
	if err != nil {
		return nil, err
	}

	err = verifyInventory(inventory)
	if err != nil {
		return nil, err
	}
	return
}
