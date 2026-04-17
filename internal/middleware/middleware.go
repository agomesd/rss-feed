package middleware

import (
	"context"

	"github.com/agomesd/rss-feed/internal/database"
	"github.com/agomesd/rss-feed/internal/handlers"
)

func LoggedIn(handler func(s *handlers.State, cmd handlers.Command, user database.User) error) func(*handlers.State, handlers.Command) error {
	return func(s *handlers.State, cmd handlers.Command) error {
		currUser, err := getCurrentUser(s)
		if err != nil {
			return err
		}
		return handler(s, cmd, currUser)
	}

}

func getCurrentUser(s *handlers.State) (database.User, error) {
	user, err := s.DB.GetUser(context.Background(), s.Config.CurrentUserName)
	if err != nil {
		return database.User{}, err
	}

	return user, err

}
