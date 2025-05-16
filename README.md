# Gator

## Program Requirements
-----------------------------------------
  - GO
  - PostgreSQL

### How to Install
-----------------------------------------
  After you have Go and PostgreSQL on your system you want to run ```go install github.com/jschasse/gator@latest```. If you DONT have your **$GOBIN** path set, set it so that you can use gator from any command line in your system. 

#### How to add the Go bin directory to your PATH (if not already there):

Identify your Go bin directory:

1. If GOBIN is set: `echo $GOBIN`
2. If GOPATH is set: `echo $GOPATH/bin`
3. Otherwise, it's likely: `$HOME/go/bin`

Add it to your PATH: Open your shell's configuration file (e.g., ~/.bashrc, ~/.zshrc):
`nano ~/.bashrc # or ~/.zshrc, etc.`

Add the following line at the end (replace /path/to/your/go/bin with the actual path from step 1):
`export PATH=$PATH:/path/to/your/go/bin`

For example, if it's $HOME/go/bin:
`export PATH=$PATH:$HOME/go/bin`

Apply the changes: Source the configuration file or open a new terminal window:
`source ~/.bashrc # or source ~/.zshrc`

Once the directory containing gator is in your PATH, you can run it from any location by simply typing:
`gator`

##### Setting up the config file

Set your database URL as follows:
`protocol://username:password@host:port/database`
Example: `postgres://admin:pw@localhost:5432/gator_db`
The filename should be `.gatorconfig.json` and located in your home directory
Inside of your `.gatorconfig.json` should look like:
> {
> "db_url": `protocol://username:password@host:port/database`
> }

Next, run the `register` command to add your user to the database.
> The `register` command takes one argument
> Example: `gator jared`
> this will register your name in the config file and set you as the user of the database.

Once you've setup the config file and registered youre name to the db you can start using the CLI to start tracking your favorite RSS feeds.

###### Commands

Commands avaliable to you:

`register <username>`: Creates a new user account with the given <username>. It also automatically logs you in as this new user.
`login <username>`: Logs you in as an existing user specified by <username>.
`users`: Lists all registered users. It will indicate which user is currently logged in.
`addfeed <feedname> <feedurl>`: Adds a new RSS feed to the system with a custom <feedname> and its <feedurl>. You must be logged in to use this. It also automatically makes the logged-in user follow this new feed.
`feeds`: Lists all the feeds that have been added to the system.
`follow <feedurl>`: Allows the logged-in user to start following an existing feed, specified by its <feedurl>.
`following`: Lists all the feeds that the currently logged-in user is following.
`unfollow <feedurl>`: Allows the logged-in user to stop following a feed, specified by its <feedurl>.
`browse [limit]`: Shows the latest posts from the feeds the logged-in user is following. You can optionally specify a [limit] (e.g., browse 5) to control how many posts are displayed (default is 2).
`agg <duration>`: Starts the aggregator service. It will continuously fetch new posts from all feeds at an interval specified by <duration> (e.g., agg 1h for every hour, agg 30m for every 30 minutes). This command will run indefinitely.
`reset`: Deletes all users from the database. This is a destructive action.
