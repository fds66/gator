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

// this command lists all the users in the database and indicates which is the current user
func handlerUsers(s *State, cmd Command) error {

	users_list, err := s.db.GetUsers(context.Background())
	if err != nil {
		fmt.Printf("Problem getting a list of users from the users database %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Users:")
	for _, name := range users_list {
		if name == s.Configuration.CurrentUserName {
			fmt.Printf("* %s (current)\n", name)
		} else {
			fmt.Printf("* %s\n", name)
		}
	}
	return nil

}

func handlerAddfeed(s *State, cmd Command) error {
	if len(cmd.Arguments) < 2 {
		return fmt.Errorf("not enough arguments provided Syntax 'addfeed name url'\n")
	}

	// create a new user in the database
	feedName := cmd.Arguments[0]
	feedURL := cmd.Arguments[1]
	//Get the current user
	currentUser := s.Configuration.CurrentUserName
	User, err := s.db.GetUser(context.Background(), currentUser)
	currentUserID := User.ID
	if err != nil {
		fmt.Printf("Cannot retrieve ID of current user %s", currentUser)
		os.Exit(1)
	}
	feedId := uuid.New()
	time := time.Now()
	createParams := database.CreateFeedParams{
		ID:        feedId,
		CreatedAt: time,
		UpdatedAt: time,
		Name:      feedName,
		Url:       feedURL,
		UserID:    currentUserID,
	}

	feed, err := s.db.CreateFeed(context.Background(), createParams)
	fmt.Printf("Feed %s at %s has been created\n", feedName, feedURL)
	fmt.Printf("Feed struct :\n")
	printFeed(feed)
	// add the feed to the current users feed follow list
	var passArgument []string
	passArgument = append(passArgument, feedURL)
	command := Command{
		Name:      "follow",
		Arguments: passArgument}
	err = handlerFollow(s, command)
	if err != nil {
		fmt.Println("Error adding this feed to the current users follow list")
	}
	return nil

}

func printFeed(feed database.Feed) {
	fmt.Printf("* ID:            %s\n", feed.ID)
	fmt.Printf("* Created:       %v\n", feed.CreatedAt)
	fmt.Printf("* Updated:       %v\n", feed.UpdatedAt)
	fmt.Printf("* Name:          %s\n", feed.Name)
	fmt.Printf("* URL:           %s\n", feed.Url)
	fmt.Printf("* UserID:        %s\n", feed.UserID)
}

// this command lists all the feeds in the database
func handlerFeeds(s *State, cmd Command) error {

	feedsList, err := s.db.ListFeeds(context.Background())
	if err != nil {
		fmt.Printf("Problem getting a list of feeds from the feeds database %v\n", err)
		os.Exit(1)
	}
	if len(feedsList) == 0 {
		fmt.Println("No feeds found")
		return nil
	}
	fmt.Printf("%d feeds found:\n", len(feedsList))
	for i := range feedsList {
		fmt.Printf("Feed %d\n", i)
		fmt.Printf("* Name:         %s\n", feedsList[i].Name)
		fmt.Printf("* URL:          %s\n", feedsList[i].Url)
		userName, err2 := s.db.UserNameFromID(context.Background(), feedsList[i].UserID)
		if err2 != nil {
			fmt.Printf("Problem retrieving username from userid for %s feed \n", feedsList[i].Name)
		}
		fmt.Printf("* created by:   %s\n", userName)
	}
	return nil

}
func handlerFollow(s *State, cmd Command) error {
	if len(cmd.Arguments) < 1 {
		return fmt.Errorf("not enough arguments provided Syntax 'follow url'\n")
	}

	// create a new follow
	// Find feed id
	feedURL := cmd.Arguments[0]
	feed, err := s.db.FeedFromURL(context.Background(), feedURL)
	if err != nil {
		fmt.Printf("Cannot retrieve ID of feed %s", feedURL)
		os.Exit(1)
	}
	feedID := feed.ID
	//fmt.Printf("feed ID %v, feedURL %v\n", feedID, feedURL)
	//Get the current user and id
	currentUser := s.Configuration.CurrentUserName
	User, err := s.db.GetUser(context.Background(), currentUser)
	currentUserID := User.ID
	if err != nil {
		fmt.Printf("Cannot retrieve ID of current user %s", currentUser)
		os.Exit(1)
	}
	//fmt.Printf("User ID %v, username %v\n", currentUserID, s.Configuration.CurrentUserName)
	feedFollowId := uuid.New()
	time := time.Now()
	createParams := database.CreateFeedFollowParams{
		ID:        feedFollowId,
		CreatedAt: time,
		UpdatedAt: time,
		UserID:    currentUserID,
		FeedID:    feedID,
	}
	//fmt.Printf("Create struct :\n")
	//fmt.Printf("%+v\n", createParams)
	feed_follow, err := s.db.CreateFeedFollow(context.Background(), createParams)
	fmt.Printf("Feed %s at %s has been created\n", feed.Name, feedURL)
	fmt.Printf("Feed struct :\n")
	fmt.Printf("%+v\n", feed_follow)
	printFeedFollow(feed_follow)

	return nil

}
func printFeedFollow(feed database.CreateFeedFollowRow) {
	fmt.Printf("* ID:            %v\n", feed.ID)
	fmt.Printf("* Created:       %v\n", feed.CreatedAt)
	fmt.Printf("* Updated:       %v\n", feed.UpdatedAt)
	fmt.Printf("* UserID:        %v\n", feed.UserID)
	fmt.Printf("* User Name:     %v\n", feed.UserName)
	fmt.Printf("* FeedID:        %v\n", feed.FeedID)
	fmt.Printf("* Feed Name:     %v\n", feed.FeedName)
}

// gets the feeds the current user is following
func handlerFollowing(s *State, cmd Command) error {

	currentUser := s.Configuration.CurrentUserName
	User, err := s.db.GetUser(context.Background(), currentUser)
	currentUserID := User.ID
	feedsList, err := s.db.GetFeedFollowsForUserID(context.Background(), currentUserID)
	if err != nil {
		fmt.Printf("Problem getting a list of feeds followed by current user %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Feeds for %s:\n", s.Configuration.CurrentUserName)
	for i := range feedsList {
		fmt.Printf("* Feed Name    %v\n", feedsList[i].FeedName)
	}
	return nil

}

// This is the reset command to remove all feedfollows from the database, useful for testing purposes
func handlerResetFeedFollow(s *State, cmd Command) error {

	err := s.db.ResetFeedFollow(context.Background())
	if err != nil {
		fmt.Printf("Problem resetting the feedfollow database %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Database feedfollow reset")
	return nil

}

// This is the reset command to remove all feeds from the database, useful for testing purposes
func handlerResetFeeds(s *State, cmd Command) error {

	err := s.db.ResetFeeds(context.Background())
	if err != nil {
		fmt.Printf("Problem resetting the feeds database %v\n", err)
		os.Exit(1)
	}
	fmt.Println("Database feeds reset")
	return nil

}

// this initiates the commands struct and registers the command names and functions
func initCommands() (Commands, error) {

	newMap := make(map[string]func(*State, Command) error)
	commands := Commands{CommandList: newMap}
	commands.register("login", handlerLogin)
	commands.register("register", handlerRegister)
	commands.register("reset", handlerReset)
	commands.register("users", handlerUsers)
	commands.register("agg", handlerAgg)
	commands.register("addfeed", handlerAddfeed)
	commands.register("feeds", handlerFeeds)
	commands.register("reset_feeds", handlerResetFeeds)
	commands.register("follow", handlerFollow)
	commands.register("following", handlerFollowing)
	commands.register("reset_feed_follow", handlerResetFeedFollow)

	return commands, nil
}
