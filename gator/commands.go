package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"

	"github.com/fds66/gator/internal/config"
	"github.com/fds66/gator/internal/database"
)

type State struct {
	db            *database.Queries
	Configuration *config.Config
}

type Command struct {
	Name      string
	Arguments []string
}

type Commands struct {
	CommandList map[string]func(*State, Command) error
}

// method to run the command
func (c *Commands) run(s *State, cmd Command) error {
	cmdName := cmd.Name
	commandFunction, exists := c.CommandList[cmdName]
	if !exists {
		return errors.New("command does not exist")
	}

	err := commandFunction(s, cmd)
	if err != nil {
		return fmt.Errorf("Error executing command, %v\n", err)
	}
	return nil
}

// method to add a new command to the commands struct
func (c *Commands) register(name string, f func(*State, Command) error) error {
	if name == "" {
		return errors.New("no name string")
	}
	c.CommandList[name] = f
	return nil

}

// this is the login function
func handlerLogin(s *State, cmd Command) error {
	if len(cmd.Arguments) == 0 {
		return fmt.Errorf("no username provided")
	}
	_, err := s.db.GetUser(context.Background(), cmd.Arguments[0])
	if err != nil {
		fmt.Printf("username %s does not exist in the database\n", cmd.Arguments[0])
		os.Exit(1)
	}
	cfg := s.Configuration
	err = cfg.SetUser(cmd.Arguments[0])
	if err != nil {
		return err
	}
	fmt.Printf("User name has been set to %s\n", cmd.Arguments[0])
	return nil
}

// this is the Register create user function
func handlerRegister(s *State, cmd Command) error {
	if len(cmd.Arguments) == 0 {
		return fmt.Errorf("no name provided")
	}
	// create a new user in the database
	username := cmd.Arguments[0]
	//Check if the username already exists
	_, err := s.db.GetUser(context.Background(), username)
	if err == nil {
		fmt.Printf("username %s already exists\n", username)
		os.Exit(1)
	}
	userId := uuid.New()
	time := time.Now()
	createParams := database.CreateUserParams{
		ID:        userId,
		CreatedAt: time,
		UpdatedAt: time,
		Name:      username}
	user := database.User{}
	user, err = s.db.CreateUser(context.Background(), createParams)
	s.Configuration.SetUser(username)
	fmt.Printf("User %s has been created\n", username)
	fmt.Printf("User struct :\n")
	fmt.Printf("id: %v\ncreated_at: %v\nupdated_at: %v\nname: %s\n", user.ID, user.CreatedAt, user.UpdatedAt, user.Name)

	return nil
}

// This is the reset command to remove all users from the database, useful for testing purposes
func handlerReset(s *State, cmd Command) error {

	err := s.db.ResetUsers(context.Background())
	if err != nil {
		fmt.Printf("Problem resetting the users database %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Database users reset")
	return nil

}

// this initiates the commands struct and registers the command names and functions
func initCommands() (Commands, error) {

	newMap := make(map[string]func(*State, Command) error)
	commands := Commands{CommandList: newMap}
	commands.register("login", handlerLogin)
	commands.register("register", handlerRegister)
	commands.register("reset", handlerReset)
	commands.register("agg", handlerAgg)

	return commands, nil
}
