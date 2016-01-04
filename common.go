package main

import (
	"fmt"
)

type ImageDefinition map[string]interface{}

type Job struct {
	Image   ImageDefinition
	Retries int
	Output  string
	Success bool
	Id      int
}

func reporter(output chan Job, done chan bool) {
	for {
		tmp := <-output
		fmt.Printf("%v", tmp.Output)
		done <- tmp.Success
	}
}
