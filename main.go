package main

import (
	"context"
	"database/sql"
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
	_ "github.com/lib/pq"
)

type state struct {
	db     *database.Queries
	config *config.Config
}

type command struct {
	name string
	args []string
}

type commands struct {
	cmds map[string]func(s *state, cmd command) error
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

func (c *commands) run(s *state, cmd command) error {
	return c.cmds[cmd.name](s, cmd)
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.cmds[name] = f
}

func handlerFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return err
	}

	for _, feed := range feeds {
		fmt.Printf("Feed: %s - From: %s - For user: %s\n", feed.Name, feed.Url, feed.Username.String)
	}

	return nil
}

func handlerReset(s *state, cmd command) error {
	return s.db.DeleteAllUsers(context.Background())
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("register command requires name")
	}

	params := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.args[0],
	}

	user, err := s.db.CreateUser(context.Background(), params)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	s.config.SetUser(user.Name)

	fmt.Println("User created:")
	fmt.Println(user)
	return nil
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("login command requires username")
	}

	username := cmd.args[0]

	user, err := s.db.GetUser(context.Background(), username)

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if err := s.config.SetUser(user.Name); err != nil {
		return err
	}

	fmt.Printf("User: %s has been set.\n", user.Name)

	return nil
}

func handlerUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return err
	}

	for _, user := range users {
		isCurrent := s.config.CurrentUserName == user.Name
		if isCurrent {
			fmt.Printf("* %s (current)\n", user.Name)
		} else {
			fmt.Printf("* %s\n", user.Name)
		}
	}

	return nil
}

func agg(s *state, cmd command) error {
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

func addFeed(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("feed name and url is required")
	}

	if len(cmd.args) == 1 {
		return fmt.Errorf("feed url is required")
	}

	user, err := s.db.GetUser(context.Background(), s.config.CurrentUserName)
	if err != nil {
		return err
	}

	params := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.args[0],
		Url:       cmd.args[1],
		UserID:    user.ID,
	}

	feed, err := s.db.CreateFeed(context.Background(), params)
	if err != nil {
		return err
	}

	fmt.Println(feed)
	return nil
}

func main() {
	c, err := config.Read()

	if err != nil {
		fmt.Println(fmt.Errorf("%w", err))
		os.Exit(1)
	}

	s := state{
		config: &c,
	}
	cmds := commands{
		cmds: make(map[string]func(s *state, cmd command) error),
	}

	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)
	cmds.register("users", handlerUsers)
	cmds.register("agg", agg)
	cmds.register("addfeed", addFeed)
	cmds.register("feeds", handlerFeeds)

	args := os.Args

	if len(args) < 2 {
		fmt.Println("command argument required")
		os.Exit(1)
	}

	cmd := command{
		name: args[1],
		args: args[2:],
	}

	db, err := sql.Open("postgres", c.DBURL)
	if err != nil {
		fmt.Println(fmt.Errorf("%w", err))
		os.Exit(1)
	}

	dbQueries := database.New(db)
	s.db = dbQueries

	err = cmds.run(&s, cmd)

	if err != nil {
		fmt.Println(fmt.Errorf("%w", err))
		os.Exit(1)
	}

	os.Exit(0)
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
