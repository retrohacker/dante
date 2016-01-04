package main

import (
	"fmt"
)

func runAlias(inventory Inventory) (errs int) {
	i := 0
	fmt.Printf("# Tagging Aliases\n\n")
	for _, image := range inventory["images"] {
		name := image["name"].(string)
		aliases := getAliasArray(image)
		for _, alias := range aliases {
			fmt.Printf("%v. %v -> %v\n", i, name, alias)
			i++
			output, err := dockerAlias(name, alias)
			if err != nil {
				fmt.Printf("Error creating tag:\n\n```\n%v\n```\n\n%v\n", output, err)
				errs++
			}
		}
	}
	fmt.Printf("\n")
	return
}

func getAliasArray(image map[string]interface{}) (aliases []string) {
	switch image["alias"].(type) {
	case string:
		// If the value is a single string, append it to the array and be done
		aliases = append(aliases, image["alias"].(string))
		break
	case []interface{}:
		// If the value is an array, iterate through and add all of the strings
		// to the array one by one.
		for _, str := range image["alias"].([]interface{}) {
			aliases = append(aliases, str.(string))
		}
		break
	}
	return
}
