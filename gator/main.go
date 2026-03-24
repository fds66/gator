package main

import (
	"fmt"
	"log"

	"github.com/fds66/gator/internal/config"
)

func main() {
	current_user_name := "fds66"
	configStruct, err := config.Read()
	if err != nil {
		log.Fatalf("error reading config from json file, %v", err)
	}
	fmt.Printf("read config, %+v\n", *configStruct)
	err = configStruct.SetUser(current_user_name)
	if err != nil {
		log.Fatalf("error setting user name, %v", err)
	}

	configStruct, err = config.Read()
	if err != nil {
		fmt.Printf("error reading data from json file, %v", err)
	}
	fmt.Printf("Read config file again: %+v\n", *configStruct)
	fmt.Printf("DbURL: '%s'\n", configStruct.DbURL)
	fmt.Printf("CurrentUserName: '%s'\n", configStruct.CurrentUserName)

}
