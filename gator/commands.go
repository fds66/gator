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

	printUser(user)

	return nil
}
func printUser(user database.User) {
	fmt.Println("User Details :")
	fmt.Printf("* ID:            %s\n", user.ID)
	fmt.Printf("* Created:       %v\n", user.CreatedAt)
	fmt.Printf("* Updated:       %v\n", user.UpdatedAt)
	fmt.Printf("* Name:          %s\n", user.Name)
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

func handlerAddFeed(s *State, cmd Command, user database.User) error {
	if len(cmd.Arguments) < 2 {
		return fmt.Errorf("not enough arguments provided Syntax 'addfeed name url'\n")
	}

	// create a new feed in the database
	feedName := cmd.Arguments[0]
	feedURL := cmd.Arguments[1]
	//Get the current user

	currentUserID := user.ID
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
	if err != nil {
		fmt.Printf("error while creating feed record, %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Feed %s at %s has been created\n", feedName, feedURL)

	printFeed(feed)
	// add the feed to the current users feed follow list
	err = CreateFeedFollowRecord(s, user, feed)
	if err != nil {
		fmt.Printf("error while creating feed follow record, %v\n", err)
		os.Exit(1)
	}

	return nil

}

func printFeed(feed database.Feed) {
	fmt.Printf("Feed Details :\n")
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
	fmt.Printf("Feeds Followed :\n")
	for i := range feedsList {
		fmt.Printf("Feed %d\n", i)
		printFeedShort(feedsList[i])
		userName, err2 := s.db.UserNameFromID(context.Background(), feedsList[i].UserID)
		if err2 != nil {
			fmt.Printf("Problem retrieving username from userid for %s feed \n", feedsList[i].Name)
		}
		fmt.Printf("* created by:   %s\n", userName)
	}
	return nil
}

func printFeedShort(feed database.Feed) {

	fmt.Printf("* Name:          %s\n", feed.Name)
	fmt.Printf("* URL:           %s\n", feed.Url)
}

func handlerFollow(s *State, cmd Command, user database.User) error {
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
	err = CreateFeedFollowRecord(s, user, feed)
	if err != nil {
		fmt.Printf("error while creating feed follow record, %v\n", err)
		os.Exit(1)
	}

	return nil

}

func CreateFeedFollowRecord(s *State, user database.User, feed database.Feed) error {
	feedFollowId := uuid.New()
	time := time.Now()
	createParams := database.CreateFeedFollowParams{
		ID:        feedFollowId,
		CreatedAt: time,
		UpdatedAt: time,
		UserID:    user.ID,
		FeedID:    feed.ID,
	}
	//fmt.Printf("Create struct :\n")
	//fmt.Printf("%+v\n", createParams)
	feed_follow, err := s.db.CreateFeedFollow(context.Background(), createParams)
	if err != nil {
		fmt.Printf("error adding record to feed follow %v\n", err)
		return err
	}
	fmt.Printf("Feed %s at %s has been created\n", feed.Name, feed.Url)

	//fmt.Printf("%+v\n", feed_follow)
	printFeedFollow(feed_follow)
	return nil
}

func printFeedFollow(feed database.CreateFeedFollowRow) {
	fmt.Println("Feed Follow Details :")
	fmt.Printf("* ID:            %v\n", feed.ID)
	fmt.Printf("* Created:       %v\n", feed.CreatedAt)
	fmt.Printf("* Updated:       %v\n", feed.UpdatedAt)
	fmt.Printf("* UserID:        %v\n", feed.UserID)
	fmt.Printf("* User Name:     %s\n", feed.UserName.String)
	fmt.Printf("* FeedID:        %v\n", feed.FeedID)
	fmt.Printf("* Feed Name:     %s\n", feed.FeedName.String)
}

// gets the feeds the current user is following
func handlerFollowing(s *State, cmd Command, user database.User) error {

	currentUserID := user.ID
	feedsList, err := s.db.GetFeedFollowsForUserID(context.Background(), currentUserID)
	if err != nil {
		fmt.Printf("Problem getting a list of feeds followed by current user %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Feeds for %s:\n", user.Name)
	for i := range feedsList {
		fmt.Printf("* Feed Name:    %s\n", feedsList[i].FeedName.String)
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
	commands.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	commands.register("feeds", handlerFeeds)
	commands.register("reset_feeds", handlerResetFeeds)
	commands.register("follow", middlewareLoggedIn(handlerFollow))
	commands.register("following", middlewareLoggedIn(handlerFollowing))
	commands.register("reset_feed_follow", handlerResetFeedFollow)

	return commands, nil
}

//Middleware
// function to check if user is logged in

func middlewareLoggedIn(handler func(s *State, cmd Command, user database.User) error) func(*State, Command) error {
	return func(s *State, cmd Command) error {
		user, err := s.db.GetUser(context.Background(), s.Configuration.CurrentUserName)
		if err != nil {
			return err
		}
		return handler(s, cmd, user)
	}

}
