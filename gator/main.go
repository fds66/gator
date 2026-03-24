package main

import (
	"fmt"

	"github.com/fds66/gator/internal/config"
)

func main() {
	current_user_name := "fds66"
	configStruct, err := config.Read()
	if err != nil {
		fmt.Printf("error reading data from json file, %v", err)
	}
	//fmt.Printf("returned config, %v\n", *configStruct)
	err2 := configStruct.SetUser(current_user_name)
	if err2 != nil {
		fmt.Printf("error setting user name, %v", err)
	}

	configStruct, err = config.Read()
	if err != nil {
		fmt.Printf("error reading data from json file, %v", err)
	}

	fmt.Printf("DbURL: '%s'\n", configStruct.DbURL)
	fmt.Printf("CurrentUserName: '%s'\n", configStruct.CurrentUserName)

}
