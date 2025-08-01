# Gator CLI

Welcome to **Gator CLI**.

You can use it to manage your own RSS feeds.

This is a Go program that uses a Postgres database, so you'll need to have **Go** and **Postgres** installed on your machine.

---

## 🛠 Installation

1. Install **Go** (if not already done)
2. Install **Postgres** (if not already done)
3. Install Gator using the following command:

   ```bash
   go install "github.com/Bigesto/gator"
   ```

---

## 🚀 Getting Started

### 1) Register a user

```bash
gator register <username>
```

> 💡 `register` will automatically set you as the current user.

---

### 2) Add RSS feeds

```bash
gator addfeed <feed_name> <feed_url>
```

> 📝 The name of the feed is up to you.  
> 📌 You’ll automatically follow the feeds you add.

---

### 3) Aggregate feed posts

```bash
gator agg <interval>
```

> ⏱ The minimal interval is `1m`.  
> `agg` runs in the background, fetching the latest posts from each feed and saving them to the database.

---

### 4) Browse followed feeds

```bash
gator browse [nb_of_posts]
```

> 📚 Displays the latest posts.  
> `nb_of_posts` is optional (default is 2).

---

## 📖 Available Commands

| Command                               | Description                                                                 |
|---------------------------------------|-----------------------------------------------------------------------------|
| `register <username>`                | Registers a new user                                                        |
| `login <username>`                   | Changes user (to an already existing one)                                   |
| `users`                              | Lists all users and highlights the active one                               |
| `addfeed <feed_name> <feed_url>`     | Adds a feed (and automatically follows it)                                  |
| `feeds`                              | Shows all existing feeds with their name, URL, and creator                  |
| `following`                          | Shows the name and URL of all feeds you are following                       |
| `follow <url>`                       | Follows an existing feed                                                    |
| `unfollow <url>`                     | Unfollows a followed feed                                                   |
| `browse [nb_of_posts]`              | Shows the latest posts (optional argument, default: 2)                      |
| `agg <interval>`                    | Aggregates new posts from followed feeds in the background (min: 1m)        |
| `reset`                              | Deletes **ALL** data from the database. A safety net is included – use wisely! |

---

🛡 **Warning:**  
`reset` will erase your entire database. This action is irreversible, so use it only if necessary.
