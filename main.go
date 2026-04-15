package main

import (
	"fmt"
	"os"

	"github.com/agomesd/rss-feed/internal/config"
)

type state struct {
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

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return fmt.Errorf("login command requires username")
	}

	username := cmd.args[0]

	if err := s.config.SetUser(username); err != nil {
		return err
	}

	fmt.Printf("User: %s has been set.\n", username)

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

	args := os.Args

	if len(args) < 2 {
		fmt.Println("command argument required")
		os.Exit(1)
	}

	cmd := command{
		name: args[1],
		args: args[2:],
	}

	err = cmds.run(&s, cmd)

	if err != nil {
		fmt.Println(fmt.Errorf("%w", err))
		os.Exit(1)
	}

	os.Exit(0)
}
