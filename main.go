package main

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
)

// === Types ===

type Item struct {
	Syntax      string
	Description string
}

type Menu []Item

// === Global Data ===

const VERSION string = "0.2.0"

var quitRequested = false
var backRequested = false
var mainMenu Menu = Menu{
	{"?            ", "Show menu"},
	{"pr           ", "Print profile"},
	{"otpauth://...", "Parse TOTP QR Code URI into profile"},
	{"ed           ", "Edit profile"},
	{"burn         ", "Burn profile to hardware token"},
	{"t            ", "Show TOTP code for profile"},
	{"q            ", "Quit"},
}
var editMenu Menu = Menu{
	{"?            ", "Show menu"},
	{"pr           ", "Print profile"},
	{"otpauth://...", "Parse TOTP QR Code URI into profile"},
	{"URI=<s>      ", "Set URI to <s> (should follow TOTP QR Code format)"},
	{"title=<s>    ", "Set title to <s> (a short hardware token profile title)"},
	{"issuer=<s>   ", "Set issuer to <s> (must not contain \":\" or \"%3A\")"},
	{"account=<s>  ", "Set account to <s> (must not contain \":\" or \"%3A\")"},
	{"secret=<s>   ", "Set secret to <s> (must be base32 string)"},
	{"algorithm=<s>", "Set algorithm to <s> (can be empty, \"SHA1\" or \"SHA256\")"},
	{"digits=<s>   ", "Set digits to <s> (can be empty, \"6\", or \"8\")"},
	{"period=<s>   ", "Set period to <s> (can be empty, \"30\", or \"60\")"},
	{"clr          ", "Clear profile"},
	{"b            ", "Go Back to main menu"},
	{"q            ", "Quit"},
}
var tmpProfile = Profile{}

// === Message Printers ===

func showMenu(m Menu) {
	for _, item := range m {
		fmt.Printf(" %v - %v\n", item.Syntax, item.Description)
	}
}

// === Input Scanners ===

var scanner *bufio.Scanner = bufio.NewScanner(os.Stdin)

func scan(prompt string) (result string) {
	fmt.Printf("%v", prompt)
	scanner.Scan()
	result = scanner.Text()
	return
}

// === Menu Action Handlers ===

func printProfile() {
	fmt.Printf("%v\n", tmpProfile)
}

func parseURI(line string) {
	tmpProfile = NewProfileFromURI(line)
	hiddenURI := tmpProfile.URI
	tmpProfile.URI = ""
	fmt.Printf("%v\n", tmpProfile)
	tmpProfile.URI = hiddenURI
}

func burnProfile() {
	printProfile()
	msg := fmt.Sprintf("Burn to token? Are you sure? [y/N]: ")
	switch scan(msg) {
	case "y", "Y":
		fmt.Println("[TODO: burn burn burn]") // TODO
	default:
		fmt.Println("[burn canceled]")
	}
}

func showTotp(p Profile) {
	t, err := NewTotp(p.Secret, p.Digits, p.Algorithm, p.Period)
	if err != nil {
		fmt.Println("Unable to show TOTP.", err)
	}
	if code, validSeconds, err := t.CurrentCode(); err != nil {
		fmt.Printf("TotpCode() error: %v\n", err)
	} else {
		fmt.Printf("%v  (%v seconds left)\n", code, validSeconds)
	}
}

// === Control Logic ===

func readEvalPrintEdit() {
	prompt := "ed> "
	line := scan(prompt)
	keyValRE := regexp.MustCompile(
		`^(URI|title|issuer|account|secret|algorithm|digits|period)=(.*)`)
	matches := keyValRE.FindStringSubmatch(line)
	n := len(matches)
	key := ""
	val := ""
	if n == 2 || n == 3 {
		key = matches[1]
	}
	if n == 3 { // Right-hand side of a key=value menu item can be blank
		val = matches[2]
	}
	switch {
	case line == "":
		// NOP
	case line == "?":
		showMenu(editMenu)
	case line == "pr":
		printProfile()
	case goodUriRE.MatchString(line):
		parseURI(line)
	case otherUriRE.MatchString(line):
		fmt.Println("URI format not recognized. Try 'ed' to enter manually?")
	case key == "URI":
		tmpProfile.URI = val
	case key == "title":
		tmpProfile.Title = val
	case key == "issuer":
		tmpProfile.Issuer = val
	case key == "account":
		tmpProfile.Account = val
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
	case line == "b":
		backRequested = true
	case line == "q!":
		quitRequested = true
	default:
		fmt.Println("Unrecognized input. Try '?' to show menu.")
	}
}

func editMenuLoop() {
	backRequested = false
	showMenu(editMenu)
	for !quitRequested && !backRequested {
		readEvalPrintEdit()
	}
	backRequested = false
}

var goodUriRE *regexp.Regexp = regexp.MustCompile(`^otpauth://totp/`)
var otherUriRE *regexp.Regexp = regexp.MustCompile(`^otpauth://`)

func readEvalPrintMain() {
	prompt := "> "
	line := scan(prompt)
	switch {
	case line == "":
		// NOP
	case line == "?":
		showMenu(mainMenu)
	case line == "pr":
		printProfile()
	case goodUriRE.MatchString(line):
		parseURI(line)
		showTotp(tmpProfile)
	case otherUriRE.MatchString(line):
		fmt.Println("URI format not recognized. Try 'ed' to enter manually?")
	case line == "ed":
		editMenuLoop()
	case line == "burn":
		burnProfile()
	case line == "t":
		showTotp(tmpProfile)
	case line == "q":
		quitRequested = true
	default:
		fmt.Println("Unrecognized input. Try '?' to show menu.")
	}
}

func main() {
	fmt.Printf("totp-util v%v\n", VERSION)
	showMenu(mainMenu)
	for !quitRequested {
		readEvalPrintMain()
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
