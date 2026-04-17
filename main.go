package main

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/agomesd/rss-feed/internal/config"
	"github.com/agomesd/rss-feed/internal/database"
	"github.com/agomesd/rss-feed/internal/handlers"
	"github.com/agomesd/rss-feed/internal/middleware"
	_ "github.com/lib/pq"
)

func main() {
	c, err := config.Read()

	if err != nil {
		fmt.Println(fmt.Errorf("%w", err))
		os.Exit(1)
	}

	s := handlers.State{
		Config: &c,
	}
	cmds := handlers.Commands{
		Cmds: make(map[string]func(s *handlers.State, cmd handlers.Command) error),
	}

	cmds.Register("login", handlers.HandlerLogin)
	cmds.Register("register", handlers.HandlerRegister)
	cmds.Register("reset", handlers.HandlerReset)
	cmds.Register("users", handlers.HandlerUsers)
	cmds.Register("agg", handlers.HandlerAgg)
	cmds.Register("addfeed", middleware.LoggedIn(handlers.HandlerAddFeed))
	cmds.Register("feeds", handlers.HandlerFeeds)
	cmds.Register("follow", middleware.LoggedIn(handlers.HandlerFollow))
	cmds.Register("following", middleware.LoggedIn(handlers.HandlerFollowing))

	args := os.Args

	if len(args) < 2 {
		fmt.Println("command argument required")
		os.Exit(1)
	}

	cmd := handlers.Command{
		Name: args[1],
		Args: args[2:],
	}

	db, err := sql.Open("postgres", c.DBURL)
	if err != nil {
		fmt.Println(fmt.Errorf("%w", err))
		os.Exit(1)
	}

	dbQueries := database.New(db)
	s.DB = dbQueries

	err = cmds.Run(&s, cmd)

	if err != nil {
		fmt.Println(fmt.Errorf("%w", err))
		os.Exit(1)
	}

	os.Exit(0)
}
