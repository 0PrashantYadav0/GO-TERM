package clipboard

import (
	"regexp"
	"time"

	"github.com/atotto/clipboard"
)

var (
	githubRepoPattern   = regexp.MustCompile(`https://github\.com/[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+`)
	npmPackagePattern   = regexp.MustCompile(`https://(?:www\.)?npmjs\.com/package/([@A-Za-z0-9_.-]+)`)
	brewPackagePattern  = regexp.MustCompile(`https://(?:www\.)?formulae\.brew\.sh/formula/([A-Za-z0-9_.-]+)`)
	urlPattern          = regexp.MustCompile(`https?://[^\s]+`)
	downloadablePattern = regexp.MustCompile(`\.(zip|pdf|jpg|jpeg|png|gif|mp3|mp4|wav|doc|docx|xls|xlsx|csv|txt|json|xml|deb)$`)
)

// Monitor continuously monitors the clipboard and sends suggestions to the channel
func Monitor(suggestions chan<- string) {
	oldText := ""

	for {
		currentText, err := clipboard.ReadAll()
		if err == nil && currentText != oldText {
			oldText = currentText
			suggestion := generateSuggestion(currentText)

			if suggestion != "" {
				suggestions <- suggestion
			}
		}

		time.Sleep(1 * time.Second)
	}
}

func generateSuggestion(text string) string {
	// Check for GitHub repository URLs
	if githubRepoPattern.MatchString(text) {
		match := githubRepoPattern.FindString(text)
		if match != "" {
			return "git clone " + match
		}
	}

	// Check for npm package URLs
	if npmPackagePattern.MatchString(text) {
		matches := npmPackagePattern.FindStringSubmatch(text)
		if len(matches) > 1 {
			return "npm install " + matches[1]
		}
	}

	// Check for Homebrew formula URLs
	if brewPackagePattern.MatchString(text) {
		matches := brewPackagePattern.FindStringSubmatch(text)
		if len(matches) > 1 {
			return "brew install " + matches[1]
		}
	}

	// Check for downloadable URLs
	if urlPattern.MatchString(text) {
		url := urlPattern.FindString(text)
		if downloadablePattern.MatchString(url) {
			return getWgetCommand(url)
		} else if url != "" {
			return "wget " + url
		}
	}

	return ""
}

func getWgetCommand(url string) string {
	filename := generateFilename()

	if regexp.MustCompile(`\.zip$`).MatchString(url) {
		return "wget -O \"" + filename + "\" \"" + url + "\" && unzip \"" + filename + "\""
	} else if regexp.MustCompile(`\.(jpg|jpeg|png|gif)$`).MatchString(url) {
		return "wget -P ./images \"" + url + "\""
	} else {
		return "wget \"" + url + "\""
	}
}

func generateFilename() string {
	now := time.Now().UnixNano()
	return "goterm_" + time.Now().Format("20060102_150405") + "_" + string(now%1000)
}

// WriteText writes text to the system clipboard
func WriteText(text string) error {
	return clipboard.WriteAll(text)
}
