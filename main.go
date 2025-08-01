package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/Bigesto/gator/internal/config"
	"github.com/Bigesto/gator/internal/database"
	_ "github.com/lib/pq"
)

func main() {
	arguments := os.Args
	configuration, err := config.Read()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	db, err := sql.Open("postgres", configuration.DbUrl)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	dbQueries := database.New(db)
	client := http.Client{}

	s := &state{
		db:     dbQueries,
		cfg:    &configuration,
		client: &client,
	}

	cmds := &commands{
		handlers: make(map[string]func(*state, command) error),
	}

	cmds.register("login", handlerLogin)
	cmds.register("register", handlerRegister)
	cmds.register("reset", handlerReset)
	cmds.register("users", handlerUsersListing)
	cmds.register("agg", handlerAgg)
	cmds.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	cmds.register("feeds", handlerFeeds)
	cmds.register("follow", middlewareLoggedIn(handlerFollow))
	cmds.register("following", middlewareLoggedIn(handlerFollowing))
	cmds.register("unfollow", middlewareLoggedIn(handlerUnfollow))
	cmds.register("browse", handlerBrowse)

	if len(arguments) < 2 {
		fmt.Println("not enough arguments, should be at least 2.")
		os.Exit(1)
	}

	toRun := command{
		name:      arguments[1],
		arguments: arguments[2:],
	}

	if err := cmds.run(s, toRun); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
