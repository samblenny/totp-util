package main

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"net/url" // For QueryUnescape
	"os"
	"regexp"
	"strings"
)

// === Types ===

type Item struct {
	Syntax      string
	Description string
}

type Menu []Item

type Profile struct {
	URI       string `json:"URI,omitempty"`
	Title     string `json:"title,omitempty"`
	Issuer    string `json:"issuer,omitempty"`
	Account   string `json:"account,omitempty"`
	Secret    string `json:"secret,omitempty"`
	Algorithm string `json:"algorithm,omitempty"`
	Digits    string `json:"digits,omitempty"`
	Period    string `json:"period,omitempty"`
}

type Database [10]Profile

// === Global Data ===

const VERSION string = "0.1.3"

var quitRequested = false
var backRequested = false
var mainMenu Menu = Menu{
	{"?            ", "Show menu"},
	{"pr           ", "Print all profiles"},
	{"otpauth://...", "Parse TOTP QR Code URI into tmp profile"},
	{"ed           ", "Edit tmp profile"},
	{"fetch <n>    ", "Fetch profile <n> into tmp for <n> in ranbe 0 to 9"},
	{"store <n>    ", "Stor tmp as profile <n> for <n> in range 0 to 9"},
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

func toJson(p Profile) string {
	bytes, err := json.MarshalIndent(p, "    ", " ")
	if err != nil {
		// This shouldn't happen
		return "{\"error\":\"something went wrong\"}"
	}
	return string(bytes)
}

func printProfile(tag string, p Profile) {
	fmt.Printf("%v %v\n", tag, strings.TrimLeft(toJson(p), ""))
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
	tmpProfile = Profile{}
	tmpProfile.URI = line
	var issuer1, account, secret, issuer2, algorithm, digits, period string
	// Remove prefix and split URI into path and query, separated by "?"
	totpQRCodeRE := regexp.MustCompile(`^otpauth://totp/(.*)\?(.*)`)
	submatches := totpQRCodeRE.FindStringSubmatch(line)
	path := submatches[1]
	// Split query into key=value pairs separated by "&"
	query := strings.Split(submatches[2], "&")
	// Split path into ((issuer)(?:$3A|:)(?:%20)*)(account=user@domain)
	pathRE := regexp.MustCompile(`(([^:]*):)(.*)`)
	pathSubmatches := pathRE.FindStringSubmatch(path)
	issuer1 = pathSubmatches[2]
	if unesc, err := url.QueryUnescape(issuer1); err == nil {
		issuer1 = unesc
	}
	account = pathSubmatches[3]
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
	if issuer1 != "" && issuer2 != "" && issuer1 != issuer2 {
		fmt.Println("WARNING: issuer prefix does not match issuer parameter")
	}
	if ok {
		// According to this wiki in the archived google-authenticator repo,
		//  https://github.com/google/google-authenticator/wiki/Key-Uri-Format
		// the recommended QR Code TOTP enrollment URI structure includes an
		// an issuer query parameter and a matching issue prefix in the label
		// that comes after `...//totp/`. But, sometimes it might be the case
		// that there's only a prefix in the label. Newer method is to use the
		// parameter but provide the label prefix for backward compatibility.
		if issuer2 != "" {
			tmpProfile.Issuer = issuer2
		} else {
			tmpProfile.Issuer = issuer1
		}
		tmpProfile.Secret = secret
		tmpProfile.Account = account
		tmpProfile.Algorithm = algorithm
		tmpProfile.Digits = digits
		tmpProfile.Period = period
		fmt.Printf("tmp %v\n", toJson(tmpProfile))
	}
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
