package main

import(
	"fmt"
	"github.com/jschasse/blogaggregator/internal/config"
)



func main() {
	c, err := config.Read()
	if err != nil {
		fmt.Print(err)
	}
	fmt.Println(c)
	c.SetUser("Jared")
	c, err = config.Read()
	if err != nil {
		fmt.Print(err)
	}
	fmt.Println(c)

}