package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"io/ioutil"
	"strconv"
)

var tempDir string

func buildImage(name string, path string) (output string, err error) {
	var outputBytes []byte

	cmd := exec.Command("docker", "build", "--no-cache", "-t", name, ".")
	cmd.Dir, err = filepath.Abs(path)
	fmt.Printf("Command: `%v`, Args: `%v`, Dir: `%v`\n",cmd.Path,cmd.Args,cmd.Dir)
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
		panic(fmt.Sprintf("Unable to get absolute path to temp directory: `%v`\n\n", err))
	}
	for _, image := range inventory["images"] {
		err = nil
		fmt.Printf("# Running `%v`\n\n## Building Image\n\n",image["name"].(string))
		output, err = buildImage(image["name"].(string), image["path"].(string))
		fmt.Printf("```\n%v\n```\n", string(output))
		if err != nil {
			fmt.Printf("**Failed** with error: `%v`\n\n",err)
			errs = append(errs, err)
			continue
		}
		//Make sure test is an array of strings, else convert it to one.
		tests := getTestArray(image)
		fmt.Printf("Array of tests: `%v`\n",tests)
		for testNum, test := range tests {
			err = nil
			var testpath string
			fmt.Printf("## Running test #%v\n\n",testNum)
			testname := image["name"].(string) + "-test" + strconv.Itoa(testNum+1)
			testpath,err = filepath.Abs(test)
			os.RemoveAll(tempDir)
			fmt.Printf("Copying `%v` to `%v`\n",testpath,tempDir)
			copyDir(testpath, tempDir)
			dockerfile := filepath.Join(tempDir, "Dockerfile")
			prependFile(dockerfile, "FROM "+image["name"].(string)+"\n")
			contents,err := ioutil.ReadFile(dockerfile)
			fmt.Printf("Contents of dockerfile `%v`:\n\n```\n%v\n```\n\n",dockerfile,string(contents))
			fmt.Printf("Building `%v` from `%v`",testname,tempDir)
			output, err = buildImage(testname, tempDir)
			fmt.Printf("```\n%v\n```\n", string(output))
			if err != nil {
				fmt.Printf("**Failed** with error: `%v`\n\n",err)
				errs = append(errs, err)
			}
		}
	}
	os.RemoveAll(tempDir)
	return
}
