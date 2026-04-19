# RSS FEED

## Instalation

You will need Postgres and Go installed to run this program - [Install Postgres](https://www.postgresql.org/download/) - [Install Go](https://go.dev/doc/install)

Install the `gator` CLI with:

```bash
go install github.com/agomesd/rss-feed
```

After installation, make sure your Go bin directory is in your PATH, then verify:

`gator`

## Configuration

1. Create a `.gatorconfif.json` file in your home directory.
2. Create a json object with 2 fields: "db_url", "current_user_name"
3. Set "db_url" to your database connection string
4. Set current_user_name to an empty string, (this will be managed by the CLI)

## Commands

- `register`: Creates a new user and sets config.current_user_name to that user. Requires username as command argument.
- `login`: Login as a user using the username. Requires username as command argument.
- `addfeed`: Adds a feed and follows that feed for the logged in user. Requires 2 command arguments: feed name and feed URL.
- `feeds`: Print all feeds.
- `follow`: Follows a feed for logged in user. Requires feed url as command argument.
- `following`: Lists all feeds being followed for the logged in user.
- `unfollow`: Unfollows a given feed for a user: Requires feed url as command argument.
- `agg`: Indefinitely runs a function to fetch posts for a user's followed feeds for a given intermitent time interval. Requires time interval as command argument ie. "1s", "1m", "1h", "1d". This command requires the user to interupt to stop it from running.
- `browse`: Lists posts for user's feeds from newsest published feeds to oldest. Takes 1 optional command argument for amount of posts to list. Defaults to 2 posts if no argument is provided.
