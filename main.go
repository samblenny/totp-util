package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"time"
)

// === Types ===

type Item struct {
	Syntax      string
	Description string
}

type Menu []Item

// === Global Data ===

const VERSION string = "0.5.0"

var quitRequested = false
var mainMenu Menu = Menu{
	{"?            ", "Show menu"},
	{"p            ", "Print profile"},
	{"otpauth://...", "Parse TOTP QR Code URI into profile"},
	{"secret=<s>   ", "Set secret to <s> (must be base32 string)"},
	{"algorithm=<s>", "Set algorithm to <s> (can be empty, \"SHA1\" or \"SHA256\")"},
	{"digits=<s>   ", "Set digits to <s> (can be empty, \"6\", or \"8\")"},
	{"period=<s>   ", "Set period to <s> (can be empty, \"30\", or \"60\")"},
	{"clr          ", "Clear profile"},
	{"t            ", "Show updating TOTP code (press Enter key to stop)"},
	{"q            ", "Quit"},
}
var tmpProfile = Profile{}

// ShowMenu prints a list of menu options
func ShowMenu(m Menu) {
	for _, item := range m {
		fmt.Printf(" %v - %v\n", item.Syntax, item.Description)
	}
}

// StdinReaderRoutine emits lines of input read from stdin
func StdinReaderRoutine(inputChan chan string, scanner *bufio.Scanner) {
	for {
		// Wait for a line of input
		scanner.Scan()
		// Toss it down the channel
		inputChan <- scanner.Text()
	}
}

// ParseURI parses a URI in the TOTP auth app QR code URI format and uses its
// query parameters to configure the current TOTP profile.
func ParseURI(line string) {
	tmpProfile = NewProfileFromURI(line)
	hiddenURI := tmpProfile.URI
	tmpProfile.URI = ""
	fmt.Printf("%v\n", tmpProfile)
	tmpProfile.URI = hiddenURI
}

// ShowTotp shows TOTP codes for the currently configured profile.
func ShowTotp(p Profile, inputChan chan string, ticker *time.Ticker) {
	t, err := NewTotp(p.Secret, p.Digits, p.Algorithm, p.Period)
	if err != nil {
		fmt.Println("Unable to show TOTP: unsupported parameter value\n", err)
		return
	}
	// Start a loop to display the TOTP code, updating every second. The loop
	// monitors the input scanner channel and stops once a line of input is
	// received.
	fmt.Printf("To stop displaying TOTP codes, use the Enter key.\n\n")
	for {
		// Block this thread until one of the channels has a message available
		select {
		case _ = <-inputChan:
			// End the loop when Enter is pressed
			fmt.Println()
			return
		case _ = <-ticker.C:
			// Generate a new code when the timer ticks
			if code, validSeconds, err := t.CurrentCode(); err != nil {
				fmt.Printf("TotpCode() error: %v\n", err)
				return
			} else {
				pad := ""
				if validSeconds < 10 {
					pad = " "
				}
				fmt.Printf("\r(%vs) %v %v  ", validSeconds, pad, code)
			}
		}
	}
}

// WaitForMenuChoice responds to inputs at the main menu prompt.
func HandleMenuChoice(inputChan chan string, ticker *time.Ticker) {
	// Get line of input from channel connected to the stdin reader goroutine
	line := <-inputChan
	// Use regular expressions to check for the more complex menu options
	goodUriRE := regexp.MustCompile(`^otpauth://totp/`)
	otherUriRE := regexp.MustCompile(`^otpauth://`)
	keyValRE := regexp.MustCompile(`^(secret|algorithm|digits|period)=(.*)`)
	matches := keyValRE.FindStringSubmatch(line)
	key := ""
	val := ""
	if len(matches) == 2 || len(matches) == 3 {
		key = matches[1]
	}
	if len(matches) == 3 { // Right-hand side of key=value can be blank
		val = matches[2]
	}
	// Match the input line against simple and complex menu options
	switch {
	case line == "":
		// NOP
	case line == "?":
		ShowMenu(mainMenu)
	case line == "p":
		fmt.Printf("%v\n", tmpProfile)
	case goodUriRE.MatchString(line):
		ParseURI(line)
		ShowTotp(tmpProfile, inputChan, ticker)
	case otherUriRE.MatchString(line):
		fmt.Println("URI format not recognized.")
	case key == "secret":
		tmpProfile.Secret = val
	case key == "algorithm":
		tmpProfile.Algorithm = val
	case key == "digits":
		tmpProfile.Digits = val
	case key == "period":
		tmpProfile.Period = val
	case line == "clr":
		tmpProfile = Profile{}
	case line == "t":
		ShowTotp(tmpProfile, inputChan, ticker)
	case line == "q":
		quitRequested = true
	default:
		fmt.Println("Unrecognized input. Try '?' to show menu.")
	}
}

func main() {
	// Show startup banner and menu options
	fmt.Printf("totp-util v%v\n", VERSION)
	ShowMenu(mainMenu)

	// Start a goroutine to read lines from stdin in a separate thread so the
	// input scanning doesn't block the updating TOTP code display on stdout
	scanner := bufio.NewScanner(os.Stdin)
	inputChan := make(chan string, 100)
	go StdinReaderRoutine(inputChan, scanner)

	// Start a 1 second tick timer
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	// Run the event loop in the main thread
	prompt := "> "
	for !quitRequested {
		fmt.Printf(prompt)
		HandleMenuChoice(inputChan, ticker)
	}
	fmt.Println("Bye")
}
