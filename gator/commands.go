package main

import (
	"errors"
	"fmt"

	"github.com/fds66/gator/internal/config"
)

type State struct {
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
	cfg := s.Configuration
	err := cfg.SetUser(cmd.Arguments[0])
	if err != nil {
		return err
	}
	fmt.Printf("User name has been set to %s\n", cmd.Arguments[0])
	return nil
}

// this initiates the commands struct and registers the command names and functions
func initCommands() (Commands, error) {

	newMap := make(map[string]func(*State, Command) error)
	commands := Commands{CommandList: newMap}
	commands.register("login", handlerLogin)

	return commands, nil
}
