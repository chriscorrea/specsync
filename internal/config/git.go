package config

import (
	"os/exec"
	"regexp"
	"strings"
)

// github repo formats
// note: narrow w/ ^ and $, shouldn't allow for appending a payload
var ownerRepoRegex = regexp.MustCompile(`^[a-zA-Z0-9][a-zA-Z0-9_.-]*/[a-zA-Z0-9][a-zA-Z0-9_.-]*$`)
var sshGitHubRegex = regexp.MustCompile(`^git@github\.com:([a-zA-Z0-9_.-]+/[a-zA-Z0-9_.-]+?)(?:\.git)?$`)
var httpsGitHubRegex = regexp.MustCompile(`^https://github\.com/([a-zA-Z0-9_.-]+/[a-zA-Z0-9_.-]+?)(?:\.git)?$`)

func ValidateOwnerRepo(s string) bool {
	if s == "" {
		return false
	}
	return ownerRepoRegex.MatchString(s)
}

// ParseGitHubRepo extracts owner/repo from a GitHub remote URL
// there might be a way to use the git library to do this?
func ParseGitHubRepo(url string) string {
	url = strings.TrimSpace(url)
	if url == "" {
		return ""
	}

	var ownerRepo string
	if matches := sshGitHubRegex.FindStringSubmatch(url); len(matches) == 2 {
		ownerRepo = matches[1]
	} else if matches := httpsGitHubRegex.FindStringSubmatch(url); len(matches) == 2 {
		ownerRepo = matches[1]
	}

	if !ValidateOwnerRepo(ownerRepo) {
		return ""
	}
	return ownerRepo
}

// DetectGitHubRepo gets owner/repo from git remote origin
func DetectGitHubRepo() string {
	cmd := exec.Command("git", "remote", "get-url", "origin")
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return ParseGitHubRepo(strings.TrimSpace(string(output)))
}
