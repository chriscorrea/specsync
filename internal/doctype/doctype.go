package doctype

import (
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

const defaultType = "document"
const scanLimit = 200 // max chars to look for keywords

var validTypes = map[string]bool{
	"specification":       true,
	"implementation_plan": true,
	"testing_plan":        true,
	"code_review":         true,
	"runbook":             true,
	"meeting":             true,
	"document":            true,
}

// keyword patterns (to be checked in order)
var keywordPatterns = []struct {
	docType  string
	keywords []string
}{
	{"specification", []string{"spec", "requirement"}},
	{"implementation_plan", []string{"implementation", "architecture", "design"}},
	{"testing_plan", []string{"testing", "test"}},
	{"code_review", []string{"review"}},
	{"runbook", []string{"runbook", "playbook"}},
	{"meeting", []string{"meeting", "standup", "retro"}},
}

var frontmatterRegex = regexp.MustCompile(`(?s)^---\n(.+?)\n---`)

// InferType infers document type from content, defaults to 'document'
func InferType(content string) string {
	// try frontmatter first
	if docType := inferFromFrontmatter(content); docType != "" {
		return docType
	}

	// try keywords in first ~200 chars
	if docType := inferFromKeywords(content); docType != "" {
		return docType
	}

	// else, return default/fallback doc type
	return defaultType
}

func inferFromFrontmatter(content string) string {
	matches := frontmatterRegex.FindStringSubmatch(content)
	if len(matches) < 2 {
		return ""
	}

	var fm struct {
		Type         string `yaml:"type"`
		DocumentType string `yaml:"document_type"`
	}

	if err := yaml.Unmarshal([]byte(matches[1]), &fm); err != nil {
		return ""
	}

	// document_type takes precedence if both present
	docType := fm.DocumentType
	if docType == "" {
		docType = fm.Type
	}

	if docType == "" {
		return ""
	}

	// validate type
	if !validTypes[docType] {
		return ""
	}

	return docType
}

func inferFromKeywords(content string) string {
	// limit scan to first N chars, defined by scanLimit above
	scanContent := content
	if len(scanContent) > scanLimit {
		scanContent = scanContent[:scanLimit]
	}
	scanContent = strings.ToLower(scanContent)

	for _, pattern := range keywordPatterns {
		for _, keyword := range pattern.keywords {
			if strings.Contains(scanContent, keyword) {
				return pattern.docType
			}
		}
	}

	return ""
}
