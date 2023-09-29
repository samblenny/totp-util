package main

import (
	"bufio"
	"errors"
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

type Database [10]Profile

// === Global Data ===

const VERSION string = "0.2.0"

var quitRequested = false
var backRequested = false
var mainMenu Menu = Menu{
	{"?            ", "Show menu"},
	{"pr           ", "Print all profiles"},
	{"otpauth://...", "Parse TOTP QR Code URI into tmp profile"},
	{"ed           ", "Edit tmp profile"},
	{"fetch <n>    ", "Fetch profile <n> into tmp for <n> in range 0 to 9"},
	{"store <n>    ", "Store tmp as profile <n> for <n> in range 0 to 9"},
	{"burn <n>     ", "Burn profile <n> to token for <n> in range 0 to 9"},
	{"q!           ", "Quit without saving changes"},
}
var editMenu Menu = Menu{
	{"?            ", "Show menu"},
	{"pr           ", "Print tmp profile"},
	{"otpauth://...", "Parse TOTP QR Code URI into tmp profile"},
	{"URI=<s>      ", "Set URI to <s> (should follow TOTP QR Code format)"},
	{"title=<s>    ", "Set title to <s> (a short hardware token profile title)"},
	{"issuer=<s>   ", "Set issuer to <s> (must not contain \":\" or \"%3A\")"},
	{"account=<s>  ", "Set account to <s> (must not contain \":\" or \"%3A\")"},
	{"secret=<s>   ", "Set secret to <s> (must be base32 string)"},
	{"algorithm=<s>", "Set algorithm to <s> (can be empty, \"SHA1\" or \"SHA256\")"},
	{"digits=<s>   ", "Set digits to <s> (can be empty, \"6\", or \"8\")"},
	{"period=<s>   ", "Set period to <s> (can be empty, \"30\", or \"60\")"},
	{"clr          ", "Clear tmp profile"},
	{"b            ", "Go Back to main menu"},
	{"q!           ", "Quit without saving changes"},
}
var profiles = Database{}
var tmpProfile = Profile{}

// === Message Printers ===

func showMenu(m Menu) {
	for _, item := range m {
		fmt.Printf(" %v - %v\n", item.Syntax, item.Description)
	}
}

func warnProfileRange(err error) {
	fmt.Println("Profile should be in range of 0 to 9.", err)
}

func warnMissingProfileArg() {
	fmt.Println("You didn't specify which profile to use. For help, try '?'.")
}

// === Input Scanners ===

var scanner *bufio.Scanner = bufio.NewScanner(os.Stdin)

func scan(prompt string) (result string) {
	fmt.Printf("%v", prompt)
	scanner.Scan()
	result = scanner.Text()
	return
}

func scanIndex(str string) (index int, err error) {
	_, err = fmt.Sscanf(str, "%d", &index)
	if err == nil && (index < 0 || 9 < index) {
		err = errors.New("Profile index out of range")
	}
	return
}

// === Menu Action Handlers ===

func printProfile(tag string, p Profile) {
	fmt.Printf("%v %v\n", tag, p)
}

func printProfileAt(index int) {
	p := profiles[index]
	tag := fmt.Sprintf("[%d]", index)
	printProfile(tag, p)
}

func printTmp() {
	printProfile("tmp", tmpProfile)
}

func printProfiles() {
	for i := range profiles {
		printProfileAt(i)
	}
	printTmp()
}

func parseURI(line string) {
	tmpProfile = NewProfileFromURI(line)
	fmt.Printf("tmp %v\n", tmpProfile)
}

func fetchProfile(index int) {
	printProfileAt(index)
	msg := fmt.Sprintf("Replace tmp with [%d]? Are you sure? [y/N]: ", index)
	switch scan(msg) {
	case "y", "Y":
		tmpProfile = profiles[index]
	default:
		fmt.Println("[fetch canceled]")
	}
}

func storeProfile(index int) {
	printProfileAt(index)
	msg := fmt.Sprintf("Replace [%d] with tmp? Are you sure? [y/N]: ", index)
	switch scan(msg) {
	case "y", "Y":
		profiles[index] = tmpProfile
		// Clear the tmp profile after storing it in a database slot. This is
		// mainly to the output of `pr` less redundant when discussing or
		// documenting console logs during testing.
		tmpProfile = Profile{}
		printTmp()
	default:
		fmt.Println("[store canceled]")
	}
}

func burnProfile(index int) {
	printProfileAt(index)
	msg := fmt.Sprintf("Burn [%d] to token? Are you sure? [y/N]: ", index)
	switch scan(msg) {
	case "y", "Y":
		fmt.Println("[TODO: burn burn burn]") // TODO
	default:
		fmt.Println("[burn canceled]")
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
		printTmp()
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

var fetchRE *regexp.Regexp = regexp.MustCompile(`^fetch [0-9]`)
var storeRE *regexp.Regexp = regexp.MustCompile(`^store [0-9]`)
var burnRE *regexp.Regexp = regexp.MustCompile(`^burn [0-9]`)
var missingRE *regexp.Regexp = regexp.MustCompile(`^(burn|set|cp)( .*$|$)`)
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
		printProfiles()
	case goodUriRE.MatchString(line):
		parseURI(line)
	case otherUriRE.MatchString(line):
		fmt.Println("URI format not recognized. Try 'ed' to enter manually?")
	case line == "ed":
		editMenuLoop()
	case fetchRE.MatchString(line):
		index, err := scanIndex(line[len("fetch "):])
		if err != nil {
			warnProfileRange(err)
		} else {
			fetchProfile(index)
		}
	case storeRE.MatchString(line):
		index, err := scanIndex(line[len("store "):])
		if err != nil {
			warnProfileRange(err)
		} else {
			storeProfile(index)
		}
	case burnRE.MatchString(line):
		index, err := scanIndex(line[len("burn "):])
		if err != nil {
			warnProfileRange(err)
		} else {
			burnProfile(index)
		}
	case missingRE.MatchString(line):
		warnMissingProfileArg()
	case line == "q!":
		quitRequested = true
	default:
		fmt.Println("Unrecognized input. Try '?' to show menu.")
	}
}

func main() {
	fmt.Println("CAUTION: Saving profiles to disk is not yet implemented!")
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
