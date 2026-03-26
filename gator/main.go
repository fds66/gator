package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/fds66/gator/internal/config"
	"github.com/fds66/gator/internal/database"
	_ "github.com/lib/pq"
)

func main() {
	// Read in config file and put into State
	configStruct, err := config.Read()
	if err != nil {
		log.Fatalf("error reading config from json file, %v", err)
	}
	//fmt.Printf("read config, %+v\n", *configStruct)

	//connection to database set up and store in state
	s := State{Configuration: configStruct}
	db, err := sql.Open("postgres", s.Configuration.DbURL)
	dbQueries := database.New(db)
	s.db = dbQueries

	// Create commands and load up
	var commands Commands
	commands, err = initCommands()
	//fmt.Printf("This is the newly created commands struct with login registered %v\n", commands)

	// get the command line arguments args[0] is the automatically the program name, args[1] is the command name, args[2 onwards] are the arguments used in the command
	args := os.Args
	if len(args) < 2 {
		fmt.Println("No command given")
		os.Exit(1)
	}
	// create the command struct
	cmd := Command{
		Name:      args[1],
		Arguments: args[2:],
	}
	fmt.Printf("command %+v\n", cmd)
	// run the command
	err = commands.run(&s, cmd)
	if err != nil {
		fmt.Println("Command returned an error")
		log.Fatalf("error executing %s, %v", cmd.Name, err)
	}
	/*
		// Read config again to check it worked
		configStruct, err = config.Read()
		if err != nil {
			fmt.Printf("error reading data from json file, %v", err)
		}
		fmt.Println("Read config file again: ")
		fmt.Printf("DbURL: '%s'\n", configStruct.DbURL)
		fmt.Printf("CurrentUserName: '%s'\n", configStruct.CurrentUserName)
	*/
}
