package main

import (
	"fmt"
	"io"
	"log"
	"strings"

	"github.com/chzyer/readline"
)

func usage(w io.Writer) {
	io.WriteString(w, "commands:\n")
	io.WriteString(w, completer.Tree("    "))
}

var completer = readline.NewPrefixCompleter(
	readline.PcItem("KEYS"),
	readline.PcItem("BUCKETS"),
	readline.PcItem("SET"),
	readline.PcItem("GET"),
	readline.PcItem("DEL"),
	readline.PcItem("BYE"),
	readline.PcItem("EXIT"),
	readline.PcItem("HELP"),
	readline.PcItem("CREATEBUCKET"),
	readline.PcItem("SETBUCKET"),
	readline.PcItem("GETBUCKET"),
)

func filterInput(r rune) (rune, bool) {
	switch r {
	// block CtrlZ feature
	case readline.CharCtrlZ:
		return r, false
	}
	return r, true
}

func client(db Database) {

	l, err := readline.NewEx(&readline.Config{
		Prompt:              "\033[31m[zombie]#\033[0m ",
		HistoryFile:         "cz.history",
		AutoComplete:        completer,
		InterruptPrompt:     "^C",
		EOFPrompt:           "exit",
		HistorySearchFold:   true,
		FuncFilterInputRune: filterInput,
	})
	if err != nil {
		panic(err)
	}
	defer l.Close()

	var bucketName string = "pages"

	log.SetOutput(l.Stderr())
	for {
		line, err := l.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}

		line = strings.TrimSpace(line)
		parts := strings.Split(line, " ")
		command := strings.ToLower(parts[0])

		switch {

		case "setbucket" == command:
			if 2 == len(parts) {
				bucketName = parts[1]
				continue
			}
			fmt.Println("Error! Incorrect usage")
			fmt.Println("SETBUCKET <bucket>")

		case "createbucket" == command:
			if 2 == len(parts) {
				db.CreateTable(parts[1])
				fmt.Println(`{"status":"ok"}`)
				continue
			}
			fmt.Println("Error! Incorrect usage")
			fmt.Println("CREATEBUCKET <bucket>")

		case "getbucket" == command:
			fmt.Println(bucketName)

		case "del" == command || "delete" == command || "rm" == command || "remove" == command:
			var key string
			if 2 == len(parts) {
				key = parts[1]
				err := db.Remove(bucketName, key)
				if nil != err {
					fmt.Println(err)
					continue
				}
				fmt.Println(`{"status":"ok"}`)
				continue
			}
			fmt.Println("Error! Incorrect usage")
			fmt.Println("DEL <key>")

		case "get" == command:
			var key string

			if 2 == len(parts) {
				key = parts[1]
				value, err := db.Get(bucketName, key)
				if nil != err {
					log.Println(err)
					continue
				}
				fmt.Println(value)
				continue
			}
			fmt.Println("Error! Incorrect usage")
			fmt.Println("GET <key>")

		case "set" == command:
			var key string
			var value string

			key = parts[1]

			i1 := strings.Index(line, "'")
			i2 := strings.LastIndex(line, "'")
			if -1 == i1 || i1 == i2 {
				fmt.Println("Error! Incorrect usage")
				fmt.Println("SET <key> '<value>'")
				continue
			}
			value = line[i1+1 : i2]

			err := db.Set(bucketName, key, value)
			if nil != err {
				fmt.Println(err)
				continue
			}
			fmt.Println(`{"status": "ok"}`)

		case command == "help":
			usage(l.Stderr())

		case "keys" == command || "ls" == command:
			results, err := db.Keys(bucketName)
			if nil != err {
				fmt.Println(err)
				continue
			}

			for i := 0; i < len(results); i++ {
				fmt.Printf("%v) %v\n", i, results[i])
			}

		case "buckets" == command:
			results, err := db.Tables()
			if nil != err {
				fmt.Println(err)
				continue
			}

			for i := 0; i < len(results); i++ {
				fmt.Printf("%v) %v\n", i+1, results[i])
			}

		case command == "bye":
			goto exit

		case command == "exit":
			goto exit

		case command == "quit":
			goto exit

		case line == "":
		default:
			fmt.Printf("Unknown command: '%v'\n", command)
		}
	}
exit:
}
