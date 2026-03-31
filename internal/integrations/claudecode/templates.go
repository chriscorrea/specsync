package claudecode

// HookScript is the bash script for the PostToolUse hook
const HookScript = `#!/bin/bash --norc

# specsync PostToolUse hook — pushes spec files after Claude edits them.
# output to stderr to avoid entering Claude's context window.

# check jq dependency
if ! command -v jq >/dev/null 2>&1; then
  echo "specsync: jq required but not found" >&2
  if command -v brew >/dev/null 2>&1; then
    echo "Install: brew install jq" >&2
  elif command -v apt-get >/dev/null 2>&1; then
    echo "Install: sudo apt-get install jq" >&2
  else
    echo "Visit https://jqlang.org/download/" >&2
  fi
  exit 0
fi

# parse the file path from hook input (JSON on stdin)
FILE_PATH=$(jq -r '.tool_input.file_path // .tool_input.path // empty')

if [ -z "$FILE_PATH" ]; then
  exit 0
fi

# reject paths with shell metacharacters
if [[ "$FILE_PATH" =~ [\;\|\&\$\` + "`" + `\(\)\{\}\<\>] ]]; then
  echo "specsync: rejected suspicious file path" >&2
  exit 0
fi

# check specsync is available
if ! command -v specsync >/dev/null 2>&1; then
  echo "specsync: not found in PATH" >&2
  exit 0
fi

# run specsync push and capture any error
ERROR=$(specsync push -- "$FILE_PATH" 2>&1 >/dev/null)
if [ $? -eq 0 ]; then
  echo "specsync: pushed $FILE_PATH" >&2
else
  echo "specsync: $FILE_PATH: $ERROR" >&2
fi

# always exit 0 — push failures should not block Claude
exit 0
`

// SettingsHook is the JSON structure for PostToolUse hook
const SettingsHook = `{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Edit|MultiEdit|Write",
        "hooks": [
          {
            "type": "command",
            "command": "\"$CLAUDE_PROJECT_DIR\"/.claude/hooks/specsync-push.sh"
          }
        ]
      }
    ]
  }
}`

// RulesContent is the markdown for .claude/rules/specsync.md
const RulesContent = `# SpecSync Integration

Spec files in this project are synced to a remote endpoint via SpecSync.
They push automatically when you write or edit them (via PostToolUse hook).

## Writing Spec Files

When creating or modifying spec/documentation files:

1. **Frontmatter recommended**: Add YAML frontmatter with a ` + "`type`" + ` field:
   ` + "```yaml" + `
   ---
   type: document
   ---
   ` + "```" + `

2. **Allowed type values**: specification, implementation_plan, testing_plan,
   code_review, runbook, architecture, meeting_notes, document

3. **File locations**: Place specs in directories matching include patterns
   (for example, ` + "`docs/**/*.md`" + ` or ` + "`specs/**/*.md`" + `)

## Conflict Handling

If you encounter ` + "`.local.md`" + `, ` + "`.remote.md`" + `, or ` + "`.base.md`" + ` files:
- These are conflict artifacts from concurrent edits; do NOT modify or delete them
`

// ClaudeMDSection is what we append to CLAUDE.md
const ClaudeMDSection = `
## SpecSync

Spec markdown files are automatically synced to a remote endpoint via SpecSync.
See ` + "`.claude/rules/specsync.md`" + ` for formatting requirements and conflict handling.
`

// HookMatcher is the regex matcher for the PostToolUse hook
const HookMatcher = "Edit|MultiEdit|Write"

// HookScriptPath is the relative path to the hook script
const HookScriptPath = ".claude/hooks/specsync-push.sh"

// SettingsPath is the relative path to settings.json
const SettingsPath = ".claude/settings.json"

// RulesPath is the relative path to the rules file
const RulesPath = ".claude/rules/specsync.md"
