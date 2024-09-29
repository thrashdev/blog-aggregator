# blog-aggregator | gator
gator is CLI tool for subscribing to RSS feeds and browsing the latest posts in the CLI.

To run this program you need an instance of PostgreSQL and the Go toolchain.

The program expect to find a config in user's home directory with the name ".gatorconfig.json".

The available commands are:
- register: registers a user, creating a record in the db
- login: logins as user, accepts a username as argument
- addfeed: adds feed to the db and adds a follow for the current user, expects a url
- follow: follows a feed, expects a url
- agg: collects posts from the feeds followed by current user
- browse: shows latest posts ordered by: feed, recency. takes an argument specifying number of posts to browse. Default is 2.
