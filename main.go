/*
   Snarky - Zero Knowledge Dead Drop
   Copyright (C) 2026 Sapadian LLC.

   This program is free software: you can redistribute it and/or modify
   it under the terms of the GNU Affero General Public License as published
   by the Free Software Foundation, either version 3 of the License, or
   (at your option) any later version.

   This program is distributed in the hope that it will be useful,
   but WITHOUT ANY WARRANTY; without even the implied warranty of
   MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
   GNU Affero General Public License for more details.

   You should have received a copy of the GNU Affero General Public License
   along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

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
	// --- Subcommands ---

	// Server Command
	serverCmd := flag.NewFlagSet("server", flag.ExitOnError)
	// We replaced the manual port flag with a config file flag
	serverConfig := serverCmd.String("config", "snarky.json", "Path to configuration file")

	// Send Command
	sendCmd := flag.NewFlagSet("send", flag.ExitOnError)
	sendFile := sendCmd.String("file", "", "File path to send")
	sendHost := sendCmd.String("host", DefaultServer, "Snarky Server URL")

	// Get Command
	getCmd := flag.NewFlagSet("get", flag.ExitOnError)
	getId := getCmd.String("id", "", "File ID")
	getKey := getCmd.String("key", "", "Decryption Key")
	getHost := getCmd.String("host", DefaultServer, "Snarky Server URL")

	// --- Argument Parsing ---

	if len(os.Args) < 2 {
		printHelp()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "server":
		serverCmd.Parse(os.Args[2:])
		// Pass the config file path to the server package
		server.Start(*serverConfig)

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
	fmt.Println("  snarky server [-config path/to/snarky.json]")
	fmt.Println("  snarky send -file <path> [-host http://...]")
	fmt.Println("  snarky get -id <ID> -key <KEY> [-host http://...]")
}
