package main

import (
	"flag"
	"fmt"
	"os"
	"snarky/client"
	"snarky/server"
)

// Default server URL - change this if you don't use the flag
const DefaultServer = "http://localhost:8080" 

func main() {
	// Subcommands
	serverCmd := flag.NewFlagSet("server", flag.ExitOnError)
	serverPort := serverCmd.String("port", "8080", "Port to run daemon on")

	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	sendFile := sendCmd.String("file", "", "File path to send")
	sendHost := sendCmd.String("host", DefaultServer, "Snarky Server URL")

	getCmd := flag.NewFlagSet("get", flag.ExitOnError)
	getId := getCmd.String("id", "", "File ID")
	getKey := getCmd.String("key", "", "Decryption Key")
	getHost := getCmd.String("host", DefaultServer, "Snarky Server URL")

	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "server":
		serverCmd.Parse(os.Args[2:])
		server.Start(*serverPort)
	case "send":
		sendCmd.Parse(os.Args[2:])
		if *sendFile == "" {
			fmt.Println("Error: -file is required")
			os.Exit(1)
		}
		client.Send(*sendHost, *sendFile)
	case "get":
		getCmd.Parse(os.Args[2:])
		if *getId == "" || *getKey == "" {
			fmt.Println("Error: -id and -key are required")
			os.Exit(1)
		}
		client.Get(*getHost, *getId, *getKey)
	default:
		printHelp()
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println("Snarky - Zero Knowledge Dead Drop")
	fmt.Println("Usage:")
	fmt.Println("  snarky server [-port 8080]")
	fmt.Println("  snarky send -file <path> [-host http://...]")
	fmt.Println("  snarky get -id <ID> -key <KEY> [-host http://...]")
}