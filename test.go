package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
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
buildImage will take a path to a docker image, and execute docker build as a
child process. It will tag the docker built image as name, this allows us to
later build other images using this one as a base. It captures stdout and
stderr returning them both in output.
*/
func buildImage(name string, path string) (output string, err error) {
	// Begin declaring local variables
	var outputBytes []byte
	// End declaring local variables

	// Build a command object that will spawn docker build as a child process.
	// We tag the the image with the provided name and don't let the build use
	// the image cache. We also set the working directory for the child process
	// to be the location of the Dockerfile
	cmd := exec.Command("docker", "build", "-t", name, ".")
	cmd.Dir, err = filepath.Abs(path)
	if err != nil {
		return
	}

	// Execute the child process and get the both stdout and stderr
	outputBytes, err = cmd.CombinedOutput()
	output = string(outputBytes)
	return
}

func buildImageNoCache(name string, path string) (output string, err error) {
	// Begin declaring local variables
	var outputBytes []byte
	// End declaring local variables

	// Build a command object that will spawn docker build as a child process.
	// We tag the the image with the provided name and don't let the build use
	// the image cache. We also set the working directory for the child process
	// to be the location of the Dockerfile
	cmd := exec.Command("docker", "build", "--no-cache", "-t", name, ".")
	cmd.Dir, err = filepath.Abs(path)
	if err != nil {
		return
	}

	// Execute the child process and get the both stdout and stderr
	outputBytes, err = cmd.CombinedOutput()
	output = string(outputBytes)
	return
}

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

func testWorker(id int, wg *sync.WaitGroup, lockOutput sync.Mutex, input chan map[string]interface{}) {
	localTempPath := tempPath + strconv.Itoa(id)
	for {
		// Get next image
		image := <-input

		// Setup local variables
		var output = fmt.Sprintf("# Tested image `%v`\n\n## Build Log\n\n", image["name"].(string))
		var resultStr string
		var err error
		var tempDir string

		// Grab an absolute path to the directory we will store our tests in.
		// We need to use a temporary directory since we will be modifying the
		// contents of the directory to build the tests against the base image.
		tempDir, err = filepath.Abs(localTempPath)
		if err != nil {
			// If we can't create a temp directory, we can't run our tests.If we can't
			// run our tests, this tool is pretty much worthless.
			panic(fmt.Sprintf("Unable to get absolute path to temp directory: `%v`\n\n", err))
		}

		// Reset error to ensure previous iterations don't polute this one
		err = nil

		// First, we build the image itself
		resultStr, err = buildImageNoCache(image["name"].(string), image["path"].(string))
		output = output + fmt.Sprintf("```\n%v\n```\n", string(resultStr))

		// If an error happens while building the image, we can't run the tests
		// against them, so we continue onto the next image.
		if err != nil {
			output = output + fmt.Sprintf("**Failed** with error: `%v`\n\n", err)
			lockOutput.Lock()
			fmt.Print(output)
			errs = append(errs, err)
			lockOutput.Unlock()
			(*wg).Done()
			continue
		}

		// Get an array of tests we want to run against our newly built image
		tests := getTestArray(image)
		output = output + fmt.Sprintf("Array of tests: `%v`\n\n", tests)

		// We will now iterate across each test building them using our newly built
		// image as a base.
		for testNum, test := range tests {
			// Reset error to ensure previous iterations don't polute this one
			err = nil
			var testpath string
			var contents []byte
			output = output + fmt.Sprintf("## Running test #%v\n\n", testNum)

			// Generate a unique name for the test image that we will build
			testname := image["name"].(string) + "-test" + strconv.Itoa(testNum+1)

			// Get the absolute path to the test Dockerfile and context location
			testpath, err = filepath.Abs(test)
			if err != nil {
				output = output + fmt.Sprintf("**Failed** Could not get path to file `%v`: `%v`\n\n", test, err)
				lockOutput.Lock()
				fmt.Print(output)
				errs = append(errs, err)
				lockOutput.Unlock()
				// If we can't get the path, we can't build the image. Moving on.
				continue
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
			output = output + fmt.Sprintf("Copying `%v` to `%v`\n", testpath, tempDir)
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
				lockOutput.Lock()
				fmt.Print(output)
				errs = append(errs, err)
				lockOutput.Unlock()
				// If we can't get the Dockerfile, we can't build the image. Moving on.
				continue
			}
			output = output + fmt.Sprintf("Contents of dockerfile `%v`:\n\n```\n%v\n```\n\n", dockerfile, string(contents))
			output = output + fmt.Sprintf("Building `%v` from `%v`", testname, tempDir)

			// Build our test image against our base image
			resultStr, err = buildImageNoCache(testname, tempDir)
			output = output + fmt.Sprintf("```\n%v\n```\n", string(resultStr))
			if err != nil {
				// If the build fails, add the error to the list of errors encountered.
				output = output + fmt.Sprintf("**Failed** with error: `%v`\n\n", err)
				lockOutput.Lock()
				fmt.Print(output)
				errs = append(errs, err)
				lockOutput.Unlock()
				continue
			}
		}
		// Remove the temp directory so we don't leave behind any trace of the tool.
		// NOTE: We are assuming the tool will run to completion. If you abort while
		// building a test, this directory will be left behind. This is something we
		// need to address moving forward.
		os.RemoveAll(tempDir)

		// Finished with this image
		lockOutput.Lock()
		fmt.Print(output)
		lockOutput.Unlock()
		(*wg).Done()
		continue
	}
}

/*
runTests iterates through an Inventory object and builds every image, followed
by running each of the tests listed against the newly built image. We attempt
to build every image defined in inventory, and return an array of errors if any
are encountered.
*/
func runTests(threads int, inventory Inventory) []error {
	// Begin declaring local variables
	var wg sync.WaitGroup
	var lockOutput sync.Mutex
	var input chan map[string]interface{} = make(chan map[string]interface{})
	// End declaring local variables

	// Start a pool of workers to handle builds
	for i := 0; i < threads; i++ {
		go testWorker(i, &wg, lockOutput, input)
	}

	// We iterate through every image, one by one, building the image and running
	// its tests
	for _, image := range inventory["images"] {
		wg.Add(1)
		input <- image
	}
	wg.Wait()
	return errs
}
