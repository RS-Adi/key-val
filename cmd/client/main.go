package main

import (
	"flag"
	"fmt"
	"os"

	"distributed-kv/pkg/client"
)

func main() {
	serverAddr := flag.String("server", "http://localhost:8080", "Server address")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Usage: client [command] [args...]")
		fmt.Println("Commands: set <key> <value>, get <key>, delete <key>")
		os.Exit(1)
	}

	c := client.NewClient(*serverAddr)
	cmd := args[0]

	switch cmd {
	case "set":
		if len(args) != 3 {
			fmt.Println("Usage: set <key> <value>")
			os.Exit(1)
		}
		if err := c.Set(args[1], args[2]); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("OK")
	case "get":
		if len(args) != 2 {
			fmt.Println("Usage: get <key>")
			os.Exit(1)
		}
		val, err := c.Get(args[1])
		if err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(val)
	case "delete":
		if len(args) != 2 {
			fmt.Println("Usage: delete <key>")
			os.Exit(1)
		}
		if err := c.Delete(args[1]); err != nil {
			fmt.Printf("Error: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("OK")
	default:
		fmt.Printf("Unknown command: %s\n", cmd)
		os.Exit(1)
	}
}
