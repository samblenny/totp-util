package main

import (
	"fmt"
	"strings"
)

// === Types ===

type Item struct {
	Shortcut    string
	Description string
}

type Menu []Item

// === Global Data ===

const VERSION string = "0.1.0"

var quitRequested = false
var unwrittenChanges = false
var dbIsOpen = false
var havePassphrase = false
var mainMenu Menu = Menu{
	{"?", "Show menu"},
	{"o", "Open database from ./totp-db.pem"},
	{"c", "Create new database in RAM"},
	{"p", "Print records"},
	{"a", "Append new record"},
	{"d", "Delete record"},
	{"e", "Edit record"},
	{"b", "Burn records to hardware token"},
	{"w", "Write database to ./totp-db.pem"},
	{"q", "Quit"},
}

// === Message Printers ===

func showMenu() {
	for _, item := range mainMenu {
		fmt.Printf(" %v - %v\n", item.Shortcut, item.Description)
	}
}

func warnDbNotOpen() {
	fmt.Println("Database is not ready. Try 'o' (open) or 'c' (create).")
}

func warnDbOpen() {
	if unwrittenChanges {
		fmt.Println("Database is already in RAM. To discard changes, try 'q'.")
	} else {
		fmt.Println("Database is already in RAM.")
	}
}

// === Input Scanners ===

func scan(prompt string) (result string) {
	fmt.Printf("%v", prompt)
	fmt.Scanf("%s", &result)
	return
}

func scanPassphrase(confirm bool) (password string) {
	fmt.Println("[prompt for passphrase]")
	if confirm {
		fmt.Println("[prompt to confirm passphrase]")
	}
	password = ""
	return
}

// === Menu Action Handlers ===

func openAndDecryptDbFile() {
	if dbIsOpen && unwrittenChanges {
		warnDbOpen()
	} else {
		_ = scanPassphrase(false)
		havePassphrase = true
		fmt.Println("[open database file ./totp-db.pem]") // TODO
		dbIsOpen = true
	}
}

// createRamDb initializes a new (or temporary) database in RAM.
// This is meant to support two use cases:
//  1. Create a new database that you intend to write to a file later using
//     totp-util's encryption facilities
//  2. Create a temporary database, ramdisk style, that you intend to use for
//     burning secrets to a hardware token without saving to disk (perhaps you
//     prefer to keep TOTP secrets in a separate password manager)
func createRamDb() {
	if dbIsOpen {
		warnDbOpen()
		return
	}
	dbIsOpen = true
	fmt.Println("[create new database in RAM]") // TODO
}

func printRecord() {
	if !dbIsOpen {
		warnDbNotOpen()
		return
	}
	fmt.Println("[print records]") // TODO
}

func appendRecord() {
	if !dbIsOpen {
		warnDbNotOpen()
		return
	}
	unwrittenChanges = true
	fmt.Println("[append record form]") // TODO
}

func deleteRecord() {
	if !dbIsOpen {
		warnDbNotOpen()
		return
	}
	unwrittenChanges = true
	fmt.Println("[delete record]") // TODO
}

func editRecord() {
	if !dbIsOpen {
		warnDbNotOpen()
		return
	}
	unwrittenChanges = true
	fmt.Println("[edit record]") // TODO
}

func burnRecordsToToken() {
	if !dbIsOpen {
		warnDbNotOpen()
		return
	}
	fmt.Println("[burn to hardware token]") // TODO
}

func writeEncryptedRecordsToDb() {
	if !dbIsOpen {
		warnDbNotOpen()
		return
	}
	if !havePassphrase {
		_ = scanPassphrase(true)
	} else {
		fmt.Println("[prompt: reuse cached passphrase?]") // TODO
	}
	fmt.Println("[encrypting db]")                   // TODO
	fmt.Println("[write database to ./totp-db.pem]") // TODO
	unwrittenChanges = false
}

func quit() {
	prompt1 := "You have unwritten database changes. Quit anyway? [y/N]: "
	prompt2 := "Last chance... You will lose data. Are you sure? [y/N]: "
	prompt3 := "To save your changes, try 'w'."
	if unwrittenChanges {
		switch strings.ToLower(scan(prompt1)) {
		case "y":
			switch strings.ToLower(scan(prompt2)) {
			case "y":
				quitRequested = true
			default:
				fmt.Println(prompt3)
			}
		default:
			fmt.Println(prompt3)
		}
	} else {
		quitRequested = true
	}
}

// === Control Logic ===

func readEvalPrint() {
	prompt := "> "
	if unwrittenChanges {
		prompt = "*> "
	}
	switch strings.ToLower(scan(prompt)) {
	case "?", "h", "help", "m", "menu":
		showMenu()
	case "o":
		openAndDecryptDbFile()
	case "c":
		createRamDb()
	case "p":
		printRecord()
	case "a":
		appendRecord()
	case "d":
		deleteRecord()
	case "e":
		editRecord()
	case "b":
		burnRecordsToToken()
	case "w":
		writeEncryptedRecordsToDb()
	case "q":
		quit()
	default:
		fmt.Println("Unrecognized input. Try '?' to show menu.")
	}
}

func main() {
	fmt.Printf("totp-util v%v\n", VERSION)
	showMenu()
	for !quitRequested {
		readEvalPrint()
	}
	fmt.Println("Bye")
}

/*
Plans:

1. File format & extension: Maybe use .pem ?
   https://en.wikipedia.org/wiki/Privacy-Enhanced_Mail
   https://www.rfc-editor.org/rfc/rfc1422

2. Encrypt database with passcode -> argon2id -> chacha20poly1305

2. Read passwords without local echo:
   https://pkg.go.dev/golang.org/x/term#ReadPassword

*/
