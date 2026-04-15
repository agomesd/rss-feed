package main

import (
	"context"
	"database/sql"
	"fmt"
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

func (c *commands) run(s *state, cmd command) error {
	return c.cmds[cmd.name](s, cmd)
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.cmds[name] = f
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
