package main

import _ "github.com/lib/pq"

import(
	"fmt"
	"github.com/jschasse/blogaggregator/internal/config"
	"github.com/jschasse/blogaggregator/internal/database"
	"database/sql"
	"os"
	"context"
	"github.com/google/uuid"
	"time"
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

	err = myCommands.run(s, co)
	if err != nil {
		fmt.Print(err)
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
	fmt.Printf("%s", user)

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