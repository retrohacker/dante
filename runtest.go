package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
)

var tempDir string

func buildImage(name string, path string) (output string, err error) {
	var outputBytes []byte

	cmd := exec.Command("docker", "build", "-t", name, ".")
	cmd.Dir, err = filepath.Abs(path)
	if err != nil {
		return
	}
	outputBytes, err = cmd.CombinedOutput()
	output = string(outputBytes)
	return
}

func getTestArray(image map[string]interface{}) (tests []string) {
	switch image["test"].(type) {
	case string:
		tests = append(tests, image["test"].(string))
		break
	case []interface{}:
		for _, str := range image["test"].([]interface{}) {
			tests = append(tests, str.(string))
		}
		break
	}
	return
}

func runTests(inventory Inventory) (errs []error) {
	var output string
	var err error
	tempDir, err = filepath.Abs(".~tmp.test")
	if err != nil {
		panic(fmt.Sprintf("Unable to get absolute path to temp directory: %v\n", err))
	}
	for _, image := range inventory["images"] {
		err = nil
		output, err = buildImage(image["name"].(string), image["path"].(string))
		errs = append(errs, err)
		fmt.Printf("%v", string(output))
		if err != nil {
			continue
		}
		//Make sure test is an array of strings, else convert it to one.
		tests := getTestArray(image)
		for testNum, test := range tests {
			err = nil
			testname := image["name"].(string) + "-test" + strconv.Itoa(testNum+1)
			fmt.Printf("%v\n", testname)
			os.RemoveAll(tempDir)
			copyDir(test, tempDir)
			dockerfile := filepath.Join(tempDir, "Dockerfile")
			prependFile(dockerfile, "FROM "+image["name"].(string)+"\n")
			fmt.Printf("%v", string(output))
			output, err = buildImage(testname, test)
			if err != nil {
				errs = append(errs, err)
			}
		}
	}
	os.RemoveAll(tempDir)
	return
}
