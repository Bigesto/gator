Welcome to Gator CLI.

You can use it to manage you own RSS feeds.

This is a Go program that uses a Postgres database, so you'll need to have Go and Postgres installed on your machine.

Installation:
- Install Go (if not already done)
- Install Postgres (if not already done)
- Install gator using the following command: go install "github.com/Bigesto/gator"

Getting started:
1) Register a user:
    gator register <username>
Note: register will automatically set you as the current user.

2) Look for some feeds and add them to your database:
    gator addfeed <feed_name> <feed_url>
Note: the name of the feed is up to you.
Note bis: you'll automatically follow the feeds you add.

3) To check the posts on each feeds you first need to aggregate them:
    gator agg <interval>
Note: the minimal time is "1m". Agg will run in the background, fetching the latest posts from each feed and saving them to the database.

4) Browse your followed feeds:
    gator browse [nb_of_posts]
Note: Shows the latest posts, nb_of_posts is optional.

Below is a list of all the available commands and their arguments:
register <username> --> registers a new user
login <username> --> changes user (to an already existing one)
users --> lists all the users (precise who is the active one)
addfeed <feed_name> <feed_url> --> adds a feed (and automatically follow it)
feeds --> shows all the existing feeds with useful data (name, url, creator)
following --> shows the name and URL of all the feeds you are following.
follow <url> --> follows an existing feed.
unfollow <url> --> unfollows a followed feed.
browse <nb_of_posts> --> shows the informations about the posts. nb_of_posts is optional (default: 2).
agg <interval> --> runs in the back and look over the feeds to add the latest posts. Minimum 1m between requests.
reset --> deletes ALL the database. You might want to consider it, so I added a safety net.