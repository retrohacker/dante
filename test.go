/*
test.go contains all of the logic specific to the test command
*/

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
)

/*
tempPath is the location that all test files will be placed in when running
this application.

NOTE: Assumption that this is a safe path in the current working directory.
This application will delete this directory when run, which is a high risk for
deleting user data. If this assumption proves not to be safe, we will have to
rethink this constant.
*/
const tempPath = ".~tmp.test"

var errs []error

/*
getTestArray takes a single image from the inventory.yml file and converts
its test key (of type interface{}) to an array of strings. This allows us
to accept either a single string or an array of strings as a value for
test.
*/
func getTestArray(image map[string]interface{}) (tests []string) {
	switch image["test"].(type) {
	case string:
		// If the value is a single string, append it to the array and be done
		tests = append(tests, image["test"].(string))
		break
	case []interface{}:
		// If the value is an array, iterate through and add all of the strings
		// to the array one by one.
		for _, str := range image["test"].([]interface{}) {
			tests = append(tests, str.(string))
		}
		break
	}
	return
}

type TestOpts struct {
	Threads int
	Retries int
}

type BuildJob struct {
	Image   ImageDefinition
	Retries int
	Output  string
	Success bool
	Id      int
}

type ImageDefinition map[string]interface{}

/*
runTests iterates through an Inventory object and builds every image, followed
by running each of the tests listed against the newly built image. We attempt
to build every image defined in inventory, and return an array of errors if any
are encountered.
*/
func runTests(inventory Inventory, opts TestOpts) (errs int) {

	input := make(chan BuildJob)
	output := make(chan BuildJob)
	done := make(chan bool, len(inventory["images"]))

	for i := 0; i < opts.Threads; i++ {
		go testWorker(input, output)
	}

	go testReporter(output, done)

	for i, image := range inventory["images"] {
		input <- BuildJob{
			Image:   image,
			Retries: opts.Retries,
			Id:      i,
		}
	}

	errs = 0
	for i := 0; i < len(inventory["images"]); i++ {
		if <-done == false {
			errs++
		}
	}

	return
}

func testWorker(input chan BuildJob, output chan BuildJob) {
	for {
		tmp := <-input
		var resultString string

		// Initialize Output For Image
		stdout := fmt.Sprintf("# Tested image `%v`\n\n## Build Log\n\n", tmp.Image["name"].(string))

		resultString, tmp = testBuildImage(tmp)
		stdout = stdout + resultString

		// If we did not successfully build, there is nothing left to do
		if !tmp.Success {
			tmp.Output = stdout
			output <- tmp
			continue
		}

		resultString, tmp = testBuildTests(tmp)
		stdout = stdout + resultString

		tmp.Output = stdout
		output <- tmp
	}
}

func testBuildImage(tmp BuildJob) (string, BuildJob) {

	var stdout string

	// Attempt to build the image until we run out of retries
	for retries := tmp.Retries; retries >= 0; retries-- {
		// Try to build the image
		result, err := buildImage(tmp.Image["name"].(string), tmp.Image["path"].(string), DockerOpts{})
		stdout = stdout + fmt.Sprintf("```\n%v\n```\n\n", string(result))

		// If we fail, determine proper notice to log, else break out of retry loop
		if err != nil {
			stdout = stdout + fmt.Sprintf("**Failed** with error: `%v`\nRetries Remaining: %v", err, retries)
			if retries == 0 {
				stdout = stdout + fmt.Sprintf("... Moving on")
				tmp.Success = false
			}
			stdout = stdout + fmt.Sprintf("\n\n")
		} else {
			tmp.Success = true
			break
		}
	}

	return stdout, tmp
}

func testBuildTests(tmp BuildJob) (string, BuildJob) {

	// Get an array of tests we want to run against our newly built image
	tests := getTestArray(tmp.Image)
	stdout := fmt.Sprintf("Array of tests: `%v`\n\n", tests)

	for testNum, test := range tests {
		output, err := testBuildTest(tmp.Image, tmp.Id, tmp.Retries, testNum, test)
		stdout = stdout + output
		if err != nil {
			tmp.Success = false
		}
	}

	return stdout, tmp
}

func testBuildTest(image ImageDefinition, id int, retries int, testNum int, test string) (output string, err error) {

	var tempDir string

	// Grab an absolute path to the directory we will store our tests in.
	// We need to use a temporary directory since we will be modifying the
	// contents of the directory to build the tests against the base image.
	localTempPath := tempPath + strconv.Itoa(id)
	tempDir, err = filepath.Abs(localTempPath)

	var testpath string
	var contents []byte
	output = output + fmt.Sprintf("## Running test #%v\n\n", testNum)

	// Generate a unique name for the test image that we will build
	testname := image["name"].(string) + "-test" + strconv.Itoa(testNum+1)

	// Get the absolute path to the test Dockerfile and context location
	testpath, err = filepath.Abs(test)
	if err != nil {
		output = output + fmt.Sprintf("**Failed** Could not get path to file `%v`: `%v`\n\n", test, err)
		// If we can't get the path, we can't build the image. Moving on.
		return
	}

	// Delete the tempDir directory if it exists to ensure our test has a
	// clean context and isn't polluted by a corrupted previous run of
	// docker-test
	os.RemoveAll(tempDir)

	// We need to copy the test context to a temp directory. In order to use
	// our new docker image as a base, we prepend a FROM statement to the
	// test's Dockerfile. This tool should be repeatable, generating the same
	// results if run in the same environment multiple times (assuming the
	// Dockerfiles it builds are deterministic). This means we can not modify
	// the original Dockerfile in place, so we create a temporary directory
	// where we copy the test's context and are then able to safely mutate
	// the context's state to suite our tool. We delete the temp directory
	// when finished.
	output = output + fmt.Sprintf("Copying `%v` to `%v`\n\n", testpath, tempDir)
	copyDir(testpath, tempDir)

	// tmpDir should already be an absolute path. So we are now getting
	// an absolute path to the Dockerfile we just copied into our tempDir
	dockerfile := filepath.Join(tempDir, "Dockerfile")

	// We then prepend a FROM statement to our Dockerfile so that when the
	// docker daemon builds it, it would build the layers on top of the
	// the image we are attempting to test.
	prependToFile(dockerfile, "FROM "+image["name"].(string)+"\n")
	contents, err = ioutil.ReadFile(dockerfile)
	if err != nil {
		output = output + fmt.Sprintf("**Failed** Could not get contents of Dockerfile `%v`: `%v`\n\n", test, err)
		// If we can't get the Dockerfile, we can't build the image. Moving on.
		return
	}
	output = output + fmt.Sprintf("Contents of dockerfile `%v`:\n\n```\n%v\n```\n\n", dockerfile, string(contents))
	output = output + fmt.Sprintf("Building `%v` from `%v`\n\n", testname, tempDir)

	// Build our test image against our base image until we succeed or run out of retries
	for ; retries >= 0; retries-- {
		var resultStr string
		resultStr, err = buildImage(testname, tempDir, DockerOpts{})
		output = output + fmt.Sprintf("```\n%v\n```\n\n", string(resultStr))
		if err != nil {
			output = output + fmt.Sprintf("**Failed** with error: `%v`\nRetries Remaining: %v", err, retries)
			if retries == 0 {
				output = output + fmt.Sprintf("... Moving on")
			}
			output = output + fmt.Sprintf("\n\n")
		} else {
			break
		}
	}
	return
}

func testReporter(output chan BuildJob, done chan bool) {
	for {
		tmp := <-output
		fmt.Printf("%v", tmp.Output)
		done <- tmp.Success
	}
}
