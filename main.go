package main

import _ "github.com/lib/pq"

import(
	"fmt"
	"github.com/jschasse/blogaggregator/internal/config"
	"github.com/jschasse/blogaggregator/internal/database"
	"github.com/lib/pq"
	"database/sql"
	"os"
	"context"
	"github.com/google/uuid"
	"time"
	"errors"
	"strconv"
	"log"
)

type state struct {
	db *database.Queries
	configPtr *config.Config 
}

type command struct {
	name string
	arguments []string
}

type commands struct {
	list map[string]func(*state, command) error
}

func middlewareLoggedIn(loggedInHandler func(s *state, cmd command, user database.User) error) func(*state, command) error {
    return func(s *state, cmd command) error {
        if s.configPtr == nil || s.configPtr.Current_user_name == "" {
            return fmt.Errorf("no user is currently logged in. Please login or register first")
        }

        user, err := s.db.GetUserByName(context.Background(), s.configPtr.Current_user_name)
        if err != nil {
            if errors.Is(err, sql.ErrNoRows) {
                return fmt.Errorf("logged in user '%s' not found in database. Please ensure you are logged in with a valid, registered user", s.configPtr.Current_user_name)
            }
            return fmt.Errorf("error fetching logged in user '%s': %w", s.configPtr.Current_user_name, err)
        }

        return loggedInHandler(s, cmd, user)
    }
}


func main() {
	var s *state
	co := command{}

	myCommands := commands{
		list: make(map[string]func(*state, command) error),
	}

	cPtr, err := config.Read()
	if err != nil {
		fmt.Printf("%s\n", err)
	}

	s = &state{}
	s.configPtr = cPtr

	db, err := sql.Open("postgres", s.configPtr.Db_url)

	dbQueries := database.New(db)

	s.db = dbQueries

	if len(os.Args) < 2 {
		fmt.Printf("Need a command name\n")
		os.Exit(1)
	}

	

	co.name = os.Args[1]
	co.arguments = os.Args[2:]

	
	myCommands.list["login"] = handlerLogin
	myCommands.list["register"] = handlerRegister
	myCommands.list["reset"] = handlerReset
	myCommands.list["users"] = handlerUsers
	myCommands.list["agg"] = handlerAgg
	myCommands.list["feeds"] = handlerFeeds


	err = myCommands.register("addfeed", middlewareLoggedIn(handlerAddFeed))
    if err != nil {
        fmt.Printf("Error registering addfeed: %v\n", err)
        os.Exit(1)
    }

    err = myCommands.register("follow", middlewareLoggedIn(handlerFollow))
    if err != nil {
        fmt.Printf("Error registering follow: %v\n", err)
		os.Exit(1)
    }

    err = myCommands.register("following", middlewareLoggedIn(handlerFollowing))
    if err != nil {
        fmt.Printf("Error registering following: %v\n", err)
		os.Exit(1)
    }

	err = myCommands.register("unfollow", middlewareLoggedIn(handlerUnfollow))
    if err != nil {
        fmt.Printf("Error registering unfollow: %v\n", err)
		os.Exit(1)
    }

	err = myCommands.register("browse", middlewareLoggedIn(handlerBrowse))
    if err != nil {
        fmt.Printf("Error registering browse: %v\n", err)
		os.Exit(1)
    }

	err = myCommands.run(s, co)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.arguments) == 0 || cmd.arguments[0] == "" {
		return fmt.Errorf("the login handler expects a single argument, the username\n")
	}

	_, err := s.db.GetUserByName(context.Background(), cmd.arguments[0])
	if err != nil {
		return fmt.Errorf("User with name %s does not exist exists\n", cmd.arguments[0])
	}

	err = s.configPtr.SetUser(cmd.arguments[0])
	if err != nil {
		return err
	}

	fmt.Printf("The user has been set\n")
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if cmd.arguments[0] == "" || len(cmd.arguments[0]) == 0 {
		return fmt.Errorf("Register command needs a name\n")
	}

	_, err := s.db.GetUserByName(context.Background(), cmd.arguments[0])
	if err == nil {
		return fmt.Errorf("User with name %s already exists\n", cmd.arguments[0])
	} else if err != sql.ErrNoRows {
		return fmt.Errorf("Error checking if user exists")
	}

	params := database.CreateUserParams{
		ID: 	   uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:	   cmd.arguments[0],
	}

	user, err := s.db.CreateUser(context.Background(), params)
	if err != nil {
		return err
	}

	err = s.configPtr.SetUser(user.Name)
	if err != nil {
		return err
	}
	fmt.Printf("%%+v: %+v\n", user)

	return nil
}

func handlerReset(s *state, cmd command) error {
	err := s.db.DeleteUsers(context.Background())
	if err != nil {
		return err
	}

	fmt.Printf("Users successfully deleted\n")
	return nil
}

func handlerUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return err
	}

	s.configPtr, err = config.Read()
	if err != nil {
		return err
	}

	for _, name := range(users) {
		if name == s.configPtr.Current_user_name {
			fmt.Printf("%s (current)\n", name)
			continue
		}
		fmt.Printf("%s\n", name)
	}

	return nil
}

func handlerAgg(s *state, cmd command) error {
	timeBetweenRequests, err := time.ParseDuration(cmd.arguments[0])
	if err != nil {
		return err
	}

	fmt.Printf("Collecting feeds every %s...", timeBetweenRequests)

	ticker := time.NewTicker(timeBetweenRequests)

	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}

	return nil
}

func handlerAddFeed(s *state, cmd command, user database.User) error {
	if len(cmd.arguments) < 2 {
		return fmt.Errorf("commands needs name and url\n")
	}
	

	params := database.CreateFeedParams{
		ID:		   uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:	   cmd.arguments[0],
		Url:	   cmd.arguments[1],
		UserID:    user.ID,
	}

	feed, err := s.db.CreateFeed(context.Background(), params)
	if err != nil {
		return err
	}

	newCMD := cmd
	newCMD.arguments = cmd.arguments[1:]

	err = handlerFollow(s, newCMD, user)

	fmt.Println(feed)

	return nil
}

func handlerFeeds(s *state, cmd command) error {
	feeds, err := s.db.GetFeeds(context.Background())
	if err != nil {
		return err
	}

	for _, feed := range feeds {
		fmt.Println(feed)
	}

	return nil
}

func scrapeFeeds(s *state) error {
	feed, err := s.db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return err
	}

	log.Println("Found the feed")
	
	scrapeFeed(s.db, feed)

	return nil
}

func scrapeFeed(db *database.Queries, feed database.Feed) error {
	err := db.MarkFeedFetched(context.Background(), feed.ID)
	if err != nil {
		return err
	}

	httpFeed, err := fetchFeed(context.Background(), feed.Url)
	if err != nil {
		return err
	}

	for _, item := range httpFeed.Channel.Item {

		pubTime, err := parsePublishedDate(item.PubDate)
		if err != nil {
			log.Printf("%s", err)
		}

		params := database.CreatePostParams{
			ID:			 uuid.New(),
			CreatedAt:	 time.Now(),
			UpdatedAt:	 time.Now(),
			Title:		 sql.NullString{String: item.Title, Valid: item.Title != ""},
			Url:		 sql.NullString{String: item.Link, Valid: item.Link != ""},
			Description: sql.NullString{String: item.Description, Valid: item.Description != ""},
			PublishedAt: sql.NullTime{Time: pubTime, Valid: !pubTime.IsZero()},
			FeedID:		 feed.ID,
		}

		_, err = db.CreatePost(context.Background(), params)
		if err != nil {
			if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505"{
				continue
			}
			return err
		}
	}

	return nil
}

func parsePublishedDate(dateStr string) (time.Time, error) {
    formats := []string{
        time.RFC1123Z, 
        time.RFC1123,   
        time.RFC3339,   
        "2006-01-02T15:04:05-07:00",
        "2006-01-02 15:04:05",
    }
    
    for _, format := range formats {
        parsedTime, err := time.Parse(format, dateStr)
        if err == nil {
            return parsedTime, nil
        }
    }
    

    return time.Time{}, fmt.Errorf("unable to parse date: %s", dateStr)
}

func handlerFollow(s *state, cmd command, user database.User) error {
	if len(cmd.arguments) < 1 {
		return fmt.Errorf("command needs a url\n")
	}

	feed, err := s.db.GetFeedByUrl(context.Background(), cmd.arguments[0])
	if err != nil {
		return err
	}


	params := database.CreateFeedFollowParams{
		ID:		   uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:	   user.ID,
		FeedID:    feed.ID,
	};

	feedFollow, err := s.db.CreateFeedFollow(context.Background(), params)
	if err != nil {
		return err
	}

	fmt.Printf("Feed name: %s Current User: %s\n", feedFollow.FeedName, s.configPtr.Current_user_name)

	return nil
}

func handlerFollowing(s *state, cmd command, user database.User) error {

	following, err := s.db.GetFeedFollowsForUser(context.Background(), user.ID)
	if err != nil {
		return err
	}

	for i := 0; i < len(following); i++ {
		fmt.Printf("%s\n", following[i].FeedName)
	}

	return nil
}

func handlerUnfollow(s *state, cmd command, user database.User) error {
	feed, err := s.db.GetFeedByUrl(context.Background(), cmd.arguments[0])
	if err != nil {
		return err
	}

	params := database.DeleteFeedFollowParams{
		UserID:	user.ID,
		FeedID:	feed.ID,
	}

	_, err = s.db.DeleteFeedFollow(context.Background(), params)
	if err != nil {
		return err
	}

	return nil
}

func handlerBrowse(s *state, cmd command, user database.User) error {
	limit := 2
	if len(cmd.arguments) == 1 {
		if specifiedLimit, err := strconv.Atoi(cmd.arguments[0]); err == nil {
			limit = specifiedLimit
		} else {
			return fmt.Errorf("invalid limit: %w", err)
		}
	}

	posts, err := s.db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  int32(limit),
	})
	if err != nil {
		return fmt.Errorf("couldn't get posts for user: %w", err)
	}

	fmt.Printf("Found %d posts for user %s:\n", len(posts), user.Name)
	for _, post := range posts {
		fmt.Printf("%s\n", post.PublishedAt.Time.Format("Mon Jan 2"))
		fmt.Printf("--- %s ---\n", post.Title)
		fmt.Printf("    %v\n", post.Description.String)
		fmt.Printf("Link: %s\n", post.Url)
		fmt.Println("=====================================")
	}

	return nil
}

func (c *commands) run(s *state, cmd command) error {
	handlerFunc, exists := c.list[cmd.name]
	if exists {
		err := handlerFunc(s, cmd)
		if err != nil {
			return err
		}
	} else {
		return fmt.Errorf("Command does not exist")
	}
	return nil
}

func (c *commands) register(name string, f func(*state, command) error) error {
	if len(name) == 0 {
		return fmt.Errorf("Command must have a name")
	}

	c.list[name] = f

	return nil
}