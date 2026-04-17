package handlers

import (
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/agomesd/rss-feed/internal/config"
	"github.com/agomesd/rss-feed/internal/database"
	"github.com/google/uuid"
)

type State struct {
	DB     *database.Queries
	Config *config.Config
}

type Command struct {
	Name string
	Args []string
}

type Commands struct {
	Cmds map[string]func(s *State, cmd Command) error
}

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDat      string `xml:"pubDate"`
}

func (c *Commands) Run(s *State, cmd Command) error {
	return c.Cmds[cmd.Name](s, cmd)
}

func (c *Commands) Register(name string, f func(*State, Command) error) {
	c.Cmds[name] = f
}

func HandlerFeeds(s *State, cmd Command) error {
	feeds, err := s.DB.GetFeeds(context.Background())
	if err != nil {
		return err
	}

	for _, feed := range feeds {
		fmt.Printf("Feed: %s - From: %s - For user: %s\n", feed.Name, feed.Url, feed.Username.String)
	}

	return nil
}

func HandlerReset(s *State, cmd Command) error {
	return s.DB.DeleteAllUsers(context.Background())
}

func HandlerRegister(s *State, cmd Command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("register command requires name")
	}

	params := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.Args[0],
	}

	user, err := s.DB.CreateUser(context.Background(), params)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	s.Config.SetUser(user.Name)

	fmt.Println("User created:")
	fmt.Println(user)
	return nil
}

func HandlerLogin(s *State, cmd Command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("login command requires username")
	}

	username := cmd.Args[0]

	user, err := s.DB.GetUser(context.Background(), username)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := s.Config.SetUser(user.Name); err != nil {
		return err
	}

	fmt.Printf("User: %s has been set.\n", user.Name)

	return nil
}

func HandlerUsers(s *State, cmd Command) error {
	users, err := s.DB.GetUsers(context.Background())
	if err != nil {
		return err
	}

	for _, user := range users {
		isCurrent := s.Config.CurrentUserName == user.Name
		if isCurrent {
			fmt.Printf("* %s (current)\n", user.Name)
		} else {
			fmt.Printf("* %s\n", user.Name)
		}
	}

	return nil
}

func HandlerAgg(s *State, cmd Command) error {
	feed, err := fetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
	if err != nil {
		return err
	}

	fmt.Println(feed)
	return nil
}

func fetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	feed := &RSSFeed{}
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return feed, err
	}

	req.Header.Set("User-Agent", "gator")

	client := http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return feed, err
	}

	data, err := io.ReadAll(res.Body)
	if err != nil {
		return feed, err
	}
	defer res.Body.Close()

	if err = xml.Unmarshal(data, feed); err != nil {
		return feed, err
	}

	cleaned := cleanFeed(*feed)

	return &cleaned, nil

}

func HandlerAddFeed(s *State, cmd Command) error {
	if len(cmd.Args) == 0 {
		return fmt.Errorf("feed name and url is required")
	}

	if len(cmd.Args) == 1 {
		return fmt.Errorf("feed url is required")
	}

	currUser, err := getCurrentUser(s)
	if err != nil {
		return err
	}

	createFeedParams := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.Args[0],
		Url:       cmd.Args[1],
		UserID:    currUser.ID,
	}

	feed, err := s.DB.CreateFeed(context.Background(), createFeedParams)
	if err != nil {
		return err
	}

	followParams := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    currUser.ID,
		FeedID:    feed.ID,
	}

	_, err = s.DB.CreateFeedFollow(context.Background(), followParams)
	if err != nil {
		return err
	}

	fmt.Println(feed)
	return nil
}

func HandlerFollowing(s *State, cmd Command) error {
	currUser, err := getCurrentUser(s)
	if err != nil {
		return err
	}

	feeds, err := s.DB.GetFeedFollowsForUser(context.Background(), currUser.ID)
	if err != nil {
		return err
	}
	for _, feed := range feeds {
		fmt.Println(feed.FeedName)
	}
	return nil
}

func HandlerFollow(s *State, cmd Command) error {
	if len(cmd.Args) < 1 {
		fmt.Println("feed url is required")
		os.Exit(1)
	}

	feed, err := s.DB.GetFeedByURL(context.Background(), cmd.Args[0])
	if err != nil {
		return err
	}
	currUser, err := getCurrentUser(s)

	if err != nil {
		return err
	}

	params := database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    currUser.ID,
		FeedID:    feed.ID,
	}

	follow, err := s.DB.CreateFeedFollow(context.Background(), params)
	if err != nil {
		return err
	}
	fmt.Printf("Feed: %s followed by %s\n", follow.FeedName, follow.UserName)
	return nil
}

func cleanFeed(f RSSFeed) RSSFeed {
	f.Channel.Title = html.UnescapeString(f.Channel.Title)
	f.Channel.Description = html.UnescapeString(f.Channel.Description)
	newItems := []RSSItem{}
	for _, item := range f.Channel.Item {
		cleanedItem := RSSItem{
			Title:       html.UnescapeString(item.Title),
			Description: html.UnescapeString(item.Description),
			Link:        item.Link,
			PubDat:      item.PubDat,
		}
		newItems = append(newItems, cleanedItem)
	}
	f.Channel.Item = newItems
	return f
}

func getCurrentUser(s *State) (database.User, error) {
	user, err := s.DB.GetUser(context.Background(), s.Config.CurrentUserName)
	if err != nil {
		return database.User{}, err
	}

	return user, err

}
