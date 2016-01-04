package main

import (
	"fmt"
)

func runPushes(inventory Inventory, opts TestOpts) (errs int) {

	input := make(chan Job)
	output := make(chan Job)
	done := make(chan bool, len(inventory["images"]))

	for i := 0; i < opts.Threads; i++ {
		go pushWorker(input, output)
	}

	go reporter(output, done)

	for i, image := range inventory["images"] {
		input <- Job{
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

	return
}

func pushWorker(input chan Job, output chan Job) {
	for {
		job := <-input
		var resultString string

		// Initialize Output For Image
		stdout := fmt.Sprintf("# Pushed image `%v`\n\n## Push Log\n\n", job.Image["name"].(string))

		resultString, job = HandleSinglePushJob(job)

		stdout = stdout + resultString

		job.Output = stdout
		output <- job

	}
}

func HandleSinglePushJob(job Job) (string, Job) {

	var stdout string

	// Attempt to build the image until we run out of retries
	for retries := job.Retries; retries >= 0; retries-- {
		// Try to build the image
		result, err := pushImage(job.Image["name"].(string))
		stdout = stdout + fmt.Sprintf("```\n%v\n```\n\n", string(result))

		// If we fail, determine proper notice to log, else break out of retry loop
		if err != nil {
			stdout = stdout + fmt.Sprintf("**Failed** with error: `%v`\nRetries Remaining: %v", err, retries)
			if retries == 0 {
				stdout = stdout + fmt.Sprintf("... Moving on")
				job.Success = false
			}
			stdout = stdout + fmt.Sprintf("\n\n")
		} else {
			job.Success = true
			break
		}
	}

	return stdout, job
}
