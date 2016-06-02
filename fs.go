package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
)

/*
copyFile moves a file from one place in the filesystem to another

Source from https://www.socketloop.com/tutorials/golang-copy-directory-including-sub-directories-files
*/
func copyFile(source string, dest string) (err error) {
	sourcefile, err := os.Open(source)
	if err != nil {
		return err
	}

	defer sourcefile.Close()

	destfile, err := os.Create(dest)
	if err != nil {
		return err
	}

	defer destfile.Close()

	_, err = io.Copy(destfile, sourcefile)
	if err == nil {
		sourceinfo, err := os.Stat(source)
		if err == nil {
			err = os.Chmod(dest, sourceinfo.Mode())
		}
	}

	return
}

/*
copyDir recursively moves the contents of a folder from one place in the filesystem to another

Source from https://www.socketloop.com/tutorials/golang-copy-directory-including-sub-directories-files
*/
func copyDir(source string, dest string) (err error) {

	// get properties of source dir
	sourceinfo, err := os.Stat(source)
	if err != nil {
		return err
	}

	// create dest dir

	err = os.MkdirAll(dest, sourceinfo.Mode())
	if err != nil {
		return err
	}

	directory, _ := os.Open(source)

	objects, err := directory.Readdir(-1)

	for _, obj := range objects {

		sourcefilepointer := source + "/" + obj.Name()

		destinationfilepointer := dest + "/" + obj.Name()

		if obj.IsDir() {
			// create sub-directories - recursively
			err = copyDir(sourcefilepointer, destinationfilepointer)
			if err != nil {
				fmt.Println(err)
			}
		} else {
			// perform copy
			err = copyFile(sourcefilepointer, destinationfilepointer)
			if err != nil {
				fmt.Println(err)
			}
		}

	}
	return
}

/*
prependToFile takes the string text and places it at the beginning of the file
named filename. This function expects filename to be a vaild path to a file,
and uses filepath.Abs() to get the absolute path of filename.
*/
func prependToFile(filename, text string) (err error) {
	// Begin declaring local variables
	var contents []byte
	// End declaring local variables

	// Get the absolute path to filename (also ensures its a valid path)
	filename, err = filepath.Abs(filename)
	if err != nil {
		return err
	}

	// Attempt to read in the entire file's contents.
	// NOTE: we are making the assumption that Dockerfiles are a reaasonable
	// size, so it is safe to load the entire file into memory when performing
	// the prepend operation
	contents, err = ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	// We then prepend the text to the beginning of the file in memory, again
	// assuming the dockerfile is a reasonable size
	contents = append([]byte(text), contents...)

	// Finally, we write the new string back to the filesystem.
	// Note: we assume this function is being used in the context of a temporary
	// directory for the purpose of this application. Since the file is a temp
	// file and only exists for the lifetime of this application running, it is
	// safe to disregard the original permission string on the file and instead
	// set it to something we know the docker daemon can use. If this assumption
	// proves to be incorrect, we will need to modify this line.
	return ioutil.WriteFile(filename, contents, 0666)
}
