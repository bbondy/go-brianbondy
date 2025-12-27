To submit a command on startup  
`claude “Do something”`

To run claude in headless mode  
`claude -p “Do something”`

To give claude code permissions in headless mode`  
`claude -p “Do something” --dangerously-skip-permissions`

Pipe data into Claude Code  
`cat data.csv | claude -p “Analyze the data”`

Create a claude.md file automatically  
`claude /init`
Or run `/init` while in Claude

Add a line to the Claude.md  
While in Claude prefix what you want to add with `#`

Slash commands  
Define .md files in `.claude/commands` dir

Slash command arguments  
Use `$ARGUMENTS` In the slash command name

Use tab for autocomplete  
Example: Modify `@browser/ui/D<tab>`

Premature compact  
`/compact`

Clear context instead of compacting  
`/clear`

Various other tips:  
- Use git a lot early and often in conjunction with Claude
- Install the Github CLI and it will use this for its interactions
- Keep an eye on auto-compact indicator on the bottom right.
- Do a code review on a PR, but post the results only here:

Advanced permission handling for headless mode:  
`claude -p "Do something" --allowedTools "Bash,Read,Edit"`

You can also use fine-grained patterns:  

Bash with specific commands:  
`claude -p "Create a commit" --allowedTools "Bash(git commit:*),Bash(git diff:*)"`

Read/Edit with path patterns:  
`claude -p "Do something" --allowedTools "Read(src/**),Edit(docs/**)"`

WebFetch with domain restrictions:  
`claude -p "Do something" --allowedTools "WebFetch(domain:github.com)"`

Common tool names  
- Bash - Shell commands
- Read - Read files
- Edit - Modify files
- Glob - File pattern matching
- Grep - Search files

Related flags  
- `--disallowedTools` - Explicitly block specific tools
- `--permission-mode` - Set mode (plan, acceptEdits, or bypassPermissions)
