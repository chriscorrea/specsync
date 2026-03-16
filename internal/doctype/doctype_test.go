package doctype

import (
	"testing"
)

func TestInferDocType(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			name: "frontmatter type specification takes precedence over keywords",
			content: `---
type: specification
---
# Project Runbook
(misleading header but frontmatter should win)`,
			want: "specification",
		},
		{
			name: "frontmatter document_type takes precedence over type and keywords",
			content: `---
document_type: runbook
type: meeting_notes
---
# Specification 
(misleading header but frontmatter should win)`,
			want: "runbook",
		},
		{
			name: "no frontmatter, implementation keyword",
			content: `# Implementation Plan

This document describes the implementation approach.`,
			want: "implementation_plan",
		},
		{
			name: "no frontmatter, testing keyword",
			content: `# Testing Plan

How we test this.`,
			want: "testing_plan",
		},
		{
			name: "no frontmatter, meeting keyword",
			content: `# Meeting Notes

Discussed the following.`,
			want: "meeting",
		},
		{
			name: "no frontmatter, no keywords",
			content: `# Random Document

Some content here.`,
			want: "document",
		},
		{
			name: "invalid frontmatter type falls back",
			content: `---
type: invalid_type
---
# Some Content (ambiguous)`,
			want: "document",
		},
		{
			name:    "empty content",
			content: "",
			want:    "document",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := InferType(tt.content)
			if got != tt.want {
				t.Errorf("InferType() = %q, want %q", got, tt.want)
			}
		})
	}
}
