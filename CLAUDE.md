# Template - Claude instructions

## Project context

- Includes web based client (separate CLAUDE.md file in /client).
- Currently a template

## Rules

- DO NOT suggest service to service injections, unless absolutely necessary - present your reasoning if so
- Match existing repository code patterns and conventions. If you'd do it differently, suggest
- Minimize helpers in any domain/service files. If they are needed, create them in utils package.
- DO NOT create separate test files, use shared per domain/service ones.