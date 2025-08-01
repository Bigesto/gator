package main

import (
	"bufio"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/Bigesto/gator/internal/config"
	"github.com/Bigesto/gator/internal/database"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

type state struct {
	db     *database.Queries
	cfg    *config.Config
	client *http.Client
}

type command struct {
	name      string
	arguments []string
}

type commands struct {
	handlers map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	handler, ok := c.handlers[cmd.name]
	if !ok {
		return fmt.Errorf("unknown command")
	}

	return handler(s, cmd)
}

func (c *commands) register(name string, handler func(*state, command) error) {
	c.handlers[name] = handler
}

func handlerLogin(s *state, cmd command) error {
	var err error

	if len(cmd.arguments) != 1 {
		return fmt.Errorf("login handler expect one argument: username")
	}

	username := cmd.arguments[0]
	context := context.Background()

	_, err = s.db.GetUser(context, username)
	if err != nil {
		fmt.Println("user is not registered yet")
		os.Exit(1)
	}

	err = s.cfg.SetUser(username)
	if err != nil {
		return err
	}

	fmt.Println("User name set as " + username)
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.arguments) != 1 {
		return fmt.Errorf("register handler expect one argument: username")
	}

	context := context.Background()

	_, err := s.db.GetUser(context, cmd.arguments[0])
	if err == nil {
		fmt.Println("user already exists")
		os.Exit(1)
	}

	userParameters := database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.arguments[0],
	}

	user, err := s.db.CreateUser(context, userParameters)
	if err != nil {
		fmt.Printf("an error occurred: %v", err)
		os.Exit(1)
	}

	err = s.cfg.SetUser(user.Name)
	if err != nil {
		return err
	}

	fmt.Println("The following user was created")
	fmt.Printf("%v\n", user)
	return nil
}

func handlerReset(s *state, cmd command) error {
	ctx := context.Background()

	areYouSure := func(reader io.Reader) bool {
		fmt.Println("This will drop all the database, are you sure about that (y/n)?")
		scanner := bufio.NewScanner(reader)
		if scanner.Scan() {
			answer := scanner.Text()
			if answer == "y" || answer == "Y" {
				return true
			}
		}
		return false
	}

	newReader := os.Stdin

	answer := areYouSure(newReader)
	if answer == true {
		err := s.db.DeleteAllUsers(ctx)
		if err != nil {
			fmt.Printf("an error occurred: %v", err)
			os.Exit(1)
		}

		fmt.Println("Users successfully deleted.")
		return nil
	}

	fmt.Println("Good idea.")
	return nil
}

func handlerUsersListing(s *state, cmd command) error {
	ctx := context.Background()

	usersList, err := s.db.GetUsers(ctx)
	if err != nil {
		return err
	}

	if len(usersList) == 0 {
		return fmt.Errorf("there are no users in the database")
	}

	for i := 0; i < len(usersList); i++ {
		if usersList[i] == s.cfg.CurrentUserName {
			fmt.Println(usersList[i] + " (current)")
		} else {
			fmt.Println(usersList[i])
		}
	}

	return nil
}

func handlerAgg(s *state, cmd command) error {
	if len(cmd.arguments) < 1 {
		return fmt.Errorf("not enough arguments, Agg need a time between requests and at least 1 minute")
	}
	unparsedTime := cmd.arguments[0]
	timeBetweenRequests, err := time.ParseDuration(unparsedTime)
	if err != nil {
		return err
	}
	const minInterval = time.Minute
	if timeBetweenRequests < minInterval {
		return fmt.Errorf("time between request is too short, no more than one request per minute")
	}
	ctx := context.Background()

	fmt.Printf("Collecting feeds every %v\n", timeBetweenRequests)
	ticker := time.NewTicker(timeBetweenRequests)
	for ; ; <-ticker.C {
		err = scrapeFeeds(s, ctx)
		if err != nil {
			fmt.Printf("Error scraping feeds: %v\n", err)
		}
	}

	return nil
}

func scrapeFeeds(s *state, ctx context.Context) error {
	nextFeed, err := s.db.GetNextFeedToFetch(ctx)
	if err != nil {
		return err
	}

	fmt.Printf("Scraping feed: %s\n", nextFeed.Name)

	err = s.db.MarkFeedFetched(ctx, nextFeed.ID)
	if err != nil {
		if err == sql.ErrNoRows {
			fmt.Println("No feeds to fetch, waiting...")
			return nil
		}
		return err
	}

	var postParameters database.CreatePostParams

	RSSfeed, err := fetchFeed(ctx, nextFeed.Url, s.client)
	if err != nil {
		return err
	}

	for _, item := range RSSfeed.Channel.Item {
		postParameters.ID = uuid.New()
		postParameters.CreatedAt = time.Now()
		postParameters.UpdatedAt = time.Now()
		postParameters.Title = nullifyString(item.Title)
		postParameters.Url = item.Link
		postParameters.Description = nullifyString(item.Description)
		postParameters.FeedID = nextFeed.ID

		pubDate, err := parsePubDate(item.PubDate)
		if err != nil {
			fmt.Println(err)
			postParameters.PublishedAt = sql.NullTime{Valid: false}
		} else {
			postParameters.PublishedAt = nullifyTime(pubDate)
		}

		err = s.db.CreatePost(ctx, postParameters)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) && pgErr.Code == "23505" {
				fmt.Printf("post %v already created\n", postParameters.Title)
				continue
			}
			return err
		}

	}

	fmt.Println("Posts successfully created")
	return nil
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.arguments) < 2 {
		return fmt.Errorf("not enough arguments, need a name and an url")
	}
	ctx := context.Background()
	name := cmd.arguments[0]
	url := cmd.arguments[1]

	user_id := user.ID

	parameters := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      name,
		Url:       url,
		UserID:    user_id,
	}

	feed, err := s.db.CreateFeed(ctx, parameters)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return fmt.Errorf("this feed already exists, just try and type gator follow %v to follow it", url)
			}
			return err
		}
		return err
	}

	fmt.Println("Feed created:")
	fmt.Println("ID:", feed.ID)
	fmt.Println("Name:", feed.Name)
	fmt.Println("URL:", feed.Url)
	fmt.Println("UserID:", feed.UserID)
	fmt.Println("CreatedAt:", feed.CreatedAt)
	fmt.Println("UpdatedAt:", feed.UpdatedAt)

	parametersFollow := database.CreateFeedFollowsParams{
		Column1: uuid.New(),
		Column2: time.Now(),
		Column3: time.Now(),
		Column4: user_id,
		Column5: feed.ID,
	}

	feedFollows, err := s.db.CreateFeedFollows(ctx, parametersFollow)
	if err != nil {
		return err
	}

	fmt.Println("Feed successfully followed")
	fmt.Println("Feed Name: " + feedFollows.FeedName)
	fmt.Println("Username: " + feedFollows.UserName)

	return nil
}

func handlerFeeds(s *state, cmd command) error {
	ctx := context.Background()

	feeds, err := s.db.GetFeeds(ctx)
	if err != nil {
		return err
	}
	if len(feeds) == 0 {
		return fmt.Errorf("database is empty")
	}

	for i, feed := range feeds {
		user, err := s.db.GetUserbyId(ctx, feed.UserID)
		if err != nil {
			return err
		}
		fmt.Printf("Database No. %d\n", i+1)
		fmt.Println(feed.Name)
		fmt.Println(feed.Url)
		fmt.Println(user.Name)
		fmt.Println("======")
	}

	return nil
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.arguments) < 1 {
		return fmt.Errorf("not enough arguments, need an url")
	}

	url := cmd.arguments[0]
	ctx := context.Background()

	feed, err := s.db.GetFeedByUrl(ctx, url)
	if err != nil {
		return err
	}

	feedId := feed.ID
	userId := user.ID

	parameters := database.CreateFeedFollowsParams{
		Column1: uuid.New(),
		Column2: time.Now(),
		Column3: time.Now(),
		Column4: userId,
		Column5: feedId,
	}

	feedFollows, err := s.db.CreateFeedFollows(ctx, parameters)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "23505" {
				return fmt.Errorf("you already follow this feed")
			}
			return err
		}
		return err
	}

	fmt.Println("Feed successfully followed")
	fmt.Println("Feed Name: " + feedFollows.FeedName)
	fmt.Println("Username: " + feedFollows.UserName)

	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {
	ctx := context.Background()

	feedFollows, err := s.db.GetFeedFollowsForUser(ctx, nullifyUUID(user.ID))
	if err != nil {
		return err
	}

	if len(feedFollows) == 0 {
		fmt.Printf("You (%v) don't follow any feeds for now.\n", user.Name)
	}

	for _, row := range feedFollows {
		feedURL, err := s.db.GetFeedURLbyID(ctx, row.FeedID)
		if err != nil {
			return err
		}
		fmt.Println(" ")
		fmt.Println(row.FeedName)
		fmt.Println(feedURL)
		fmt.Println("================")
	}

	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	if len(cmd.arguments) < 1 {
		return fmt.Errorf("not enough arguments, needs the url of the feed to unfollow")
	}
	ctx := context.Background()

	url := cmd.arguments[0]

	parameters := database.DeleteFeedFollowsByUserAndURLParams{
		Column1: nullifyUUID(user.ID),
		Column2: nullifyString(url),
	}

	_, err := s.db.DeleteFeedFollowsByUserAndURL(ctx, parameters)
	if errors.Is(err, sql.ErrNoRows) {
		fmt.Println("The provided address is wrong, or you weren't following that feed.")
		return nil
	}
	if err != nil {
		return err
	}

	fmt.Println("Feed unfollowed successfully.")
	return nil
}

func handlerBrowse(s *state, cmd command) error {
	var limit int64
	var err error
	if len(cmd.arguments) < 1 {
		limit = 2
	} else {
		i, err := strconv.ParseInt(cmd.arguments[0], 10, 64)
		if err != nil {
			return err
		}
		limit = i
	}
	ctx := context.Background()

	user, err := s.db.GetUser(ctx, s.cfg.CurrentUserName)
	if err != nil {
		return err
	}

	feeds, err := s.db.GetFeedFollowsForUser(ctx, nullifyUUID(user.ID))
	if err != nil {
		return err
	}

	feedsID := make([]uuid.UUID, len(feeds))

	for _, feed := range feeds {
		feedsID = append(feedsID, feed.FeedID)
	}

	getPostsParam := database.GetPostsForUserParams{
		Column1: feedsID,
		Limit:   limit,
	}

	posts, err := s.db.GetPostsForUser(ctx, getPostsParam)
	if err != nil {
		return nil
	}

	for _, post := range posts {
		fmt.Printf("Title: %v\n", post.Title)
		fmt.Printf("Published At: %v\n", post.PublishedAt)
		fmt.Printf("URL: %v\n", post.Url)
		fmt.Printf("Description: %v\n", post.Description)
		fmt.Println("==================")
		fmt.Println(" ")
	}
	return nil
}

func nullifyString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

func nullifyUUID(id uuid.UUID) uuid.NullUUID {
	return uuid.NullUUID{UUID: id, Valid: true}
}

func parsePubDate(pubDate string) (time.Time, error) {
	layouts := []string{
		time.RFC1123Z, time.RFC1123, time.RFC822Z, time.RFC822, time.RFC3339,
		"2006-01-02 15:04:05", "2006-01-02",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, pubDate); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unrecognized date format: %s", pubDate)
}

func nullifyTime(t time.Time) sql.NullTime {
	return sql.NullTime{Time: t, Valid: true}
}
