package main

import (
	"bufio"
	"errors"
	"fmt"
	"net/url" // For QueryUnescape
	"os"
	"regexp"
	"strings"
)

// === Types ===

type Item struct {
	Shortcut    string
	Description string
}

type Menu []Item

type Profile struct {
	URI    string
	Title  string
	Secret string
}

type Database [10]Profile

// === Global Data ===

const VERSION string = "0.1.1"

var quitRequested = false
var mainMenu Menu = Menu{
	{"?             ", "Show menu"},
	{"pr            ", "Print all profiles"},
	{"otpauth://... ", "Parse TOTP QR Code URI into tmp profile"},
	{"ed            ", "Edit tmp profile"},
	{"get <profile> ", "Get profile into tmp for <profile> in range 0 to 9"},
	{"set <profile> ", "Set profile from tmp for <profile> in range 0 to 9"},
	{"burn <profile>", "Burn profile to token for <profile> in range 0 to 9"},
	{"q!            ", "Quit without saving changes"},
}
var profiles = Database{
	{"", "", ""}, // 0
	{"", "", ""}, // 1
	{"", "", ""}, // 2
	{"", "", ""}, // 3
	{"", "", ""}, // 4
	{"", "", ""}, // 5
	{"", "", ""}, // 6
	{"", "", ""}, // 7
	{"", "", ""}, // 8
	{"", "", ""}, // 9
}
var tmpProfile = Profile{
	"otpauth://totp/Example:alice@google.com?secret=JBSWY3DPEHPK3PXP&issuer=Example",
	"Example",
	"JBSWY3DPEHPK3PXP",
}

// === Message Printers ===

func showMenu() {
	for _, item := range mainMenu {
		fmt.Printf(" %v - %v\n", item.Shortcut, item.Description)
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
	_ = scanPassphrase(false)
	fmt.Println("[open database file ./totp-db.pem]") // TODO
}

func printProfile(index int) {
	p := profiles[index]
	fmt.Printf("[%d]    URI: %v\n", index, p.URI)
	fmt.Printf("     title: %v\n", p.Title)
	fmt.Printf("    secret: %v\n", p.Secret)
}

func printTmp() {
	fmt.Printf("tmp    URI: %v\n", tmpProfile.URI)
	fmt.Printf("     title: %v\n", tmpProfile.Title)
	fmt.Printf("    secret: %v\n", tmpProfile.Secret)
}

func printProfiles() {
	for i := range profiles {
		printProfile(i)
	}
	printTmp()
}

func parseURI(line string) {
	tmpProfile = Profile{line, "", ""}
	var issuer1, label, secret, issuer2, algorithm, digits, period string
	// Remove prefix and split URI into path and query, separated by "?"
	totpQRCodeRE := regexp.MustCompile(`^otpauth://totp/(.*)\?(.*)`)
	submatches := totpQRCodeRE.FindStringSubmatch(line)
	path := submatches[1]
	// Split query into key=value pairs separated by "&"
	query := strings.Split(submatches[2], "&")
	// Split path into ((issuer):)(label=user@domain)
	pathRE := regexp.MustCompile(`(([^:]*):)(.*)`)
	pathSubmatches := pathRE.FindStringSubmatch(path)
	issuer1 = pathSubmatches[2]
	if unesc, err := url.QueryUnescape(issuer1); err == nil {
		issuer1 = unesc
	}
	label = pathSubmatches[3]
	// Extract query parameter values (secret=, ...)
	for _, v := range query {
		switch {
		case strings.HasPrefix(v, "secret="):
			secret = v[len("secret="):]
		case strings.HasPrefix(v, "issuer="):
			issuer2 = v[len("issuer="):]
			if unesc, err := url.QueryUnescape(issuer2); err == nil {
				issuer2 = unesc
			}
		case strings.HasPrefix(v, "algorithm="):
			algorithm = strings.ToUpper(v[len("algorithm="):])
		case strings.HasPrefix(v, "digits="):
			digits = v[len("digits="):]
		case strings.HasPrefix(v, "period="):
			period = v[len("period="):]
		}
	}
	fmt.Printf("   issuer1: %v\n", issuer1)
	fmt.Printf("     label: %v\n", label)
	fmt.Printf("    secret: %v\n", secret)
	fmt.Printf("    issuer: %v\n", issuer2)
	fmt.Printf(" algorithm: %v\n", algorithm)
	fmt.Printf("    digits: %v\n", digits)
	fmt.Printf("    period: %v\n", period)
	var ok = true
	// Google Authenticator expects 6 digits, SHA1, and 30s period, so for
	// URI's that specify something different, flag them for manual review
	if algorithm != "" && algorithm != "SHA1" {
		fmt.Println("WARNING: Unsupported algorithm (not blank, not SHA1)")
		ok = false
	}
	if period != "" && period != "30" {
		fmt.Println("WARNING: Unsupported period (not blank, not 30)")
		ok = false
	}
	if digits != "" && digits != "6" {
		fmt.Println("WARNING: Unsupported digits (not blank, not 6)")
		ok = false
	}
	if ok {
		tmpProfile.Title = issuer2
		tmpProfile.Secret = secret
	}
}

func editTmp() {
	fmt.Println("[edit tmp]") // TODO
}

func getProfile(index int) {
	printProfile(index)
	printTmp()
	msg := fmt.Sprintf("Replace tmp with [%d]? Are you sure? [y/N]: ", index)
	switch scan(msg) {
	case "y", "Y":
		tmpProfile = profiles[index]
		fmt.Printf("[get: copied profile %d to tmp]\n", index)
	default:
		fmt.Println("[get canceled]")
	}
}

func setProfile(index int) {
	printProfile(index)
	printTmp()
	msg := fmt.Sprintf("Replace [%d] with tmp? Are you sure? [y/N]: ", index)
	switch scan(msg) {
	case "y", "Y":
		profiles[index] = tmpProfile
		fmt.Printf("[set profile %d to values from tmp]\n", index)
	default:
		fmt.Println("[set canceled]")
	}
}

func burnProfile(index int) {
	printProfile(index)
	msg := fmt.Sprintf("Burn [%d] to token? Are you sure? [y/N]: ", index)
	switch scan(msg) {
	case "y", "Y":
		fmt.Printf("[burn to hardware token profile %d]\n", index)
		fmt.Println("[TODO: burn burn burn]") // TODO
	default:
		fmt.Println("[burn canceled]")
	}
}

func quit() {
	quitRequested = true
}

// === Control Logic ===

var getRE *regexp.Regexp = regexp.MustCompile(`^get [0-9]`)
var setRE *regexp.Regexp = regexp.MustCompile(`^set [0-9]`)
var burnRE *regexp.Regexp = regexp.MustCompile(`^burn [0-9]`)
var missingRE *regexp.Regexp = regexp.MustCompile(`^(burn|set|cp)( .*$|$)`)
var goodUriRE *regexp.Regexp = regexp.MustCompile(`^otpauth://totp/`)
var otherUriRE *regexp.Regexp = regexp.MustCompile(`^otpauth://`)

func readEvalPrint() {
	prompt := "> "
	line := scan(prompt)
	switch {
	case line == "?":
		showMenu()
	case line == "pr":
		printProfiles()
	case goodUriRE.MatchString(line):
		parseURI(line)
	case otherUriRE.MatchString(line):
		fmt.Println("URI format not recognized. Try 'ed' to enter manually?")
	case line == "ed":
		editTmp()
	case getRE.MatchString(line):
		index, err := scanIndex(line[4:]) // len("get ") == 4
		if err != nil {
			warnProfileRange(err)
		} else {
			getProfile(index)
		}
	case setRE.MatchString(line):
		index, err := scanIndex(line[4:]) // len("set ") == 4
		if err != nil {
			warnProfileRange(err)
		} else {
			setProfile(index)
		}
	case burnRE.MatchString(line):
		index, err := scanIndex(line[5:]) // len("burn ") == 5
		if err != nil {
			warnProfileRange(err)
		} else {
			burnProfile(index)
		}
	case missingRE.MatchString(line):
		warnMissingProfileArg()
	case line == "q!":
		quit()
	default:
		fmt.Println("Unrecognized input. Try '?' to show menu.")
	}
}

func main() {
	fmt.Println("CAUTION: Saving profiles to disk is not yet implemented!")
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
