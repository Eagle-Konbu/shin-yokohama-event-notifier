# Purpose

This repository contains a small, personal Discord notification bot.

The bot scrapes event information from selected venues around the Shin-Yokohama area and sends notifications once per day.

This project prioritizes simplicity, readability, and maintainability over extensibility or general-purpose design.

---

# Scope and Constraints

- This is a small-scale, single-purpose project.
- Avoid large architectural changes.
- Do not introduce unnecessary abstractions or frameworks.
- Keep changes minimal and localized.

Not allowed:
- Over-engineering for hypothetical future use cases
- Turning this project into a reusable library
- Adding features unrelated to event notification

---

# Runtime Assumptions

- Runs on AWS Lambda.
- Stateless execution.
- Triggered once per day via Amazon EventBridge.
- Short execution time and low concurrency.

Do not assume:
- Long-running processes
- Persistent local storage
- Interactive or user-driven execution

---

# Architecture

This project adopts hexagonal architecture (ports and adapters pattern).

Structure:
- `domain/`: Core business logic and entities
- `domain/ports/`: Interfaces for external dependencies
- `application/`: Use cases and application services
- `infrastructure/`: Adapter implementations (Discord, HTTP clients, etc.)

Note: Implementation is currently in progress. Not all components follow this structure yet.

The architecture aims to:
- Isolate business logic from external dependencies
- Make the code testable without external services
- Keep dependencies pointing inward

However:
- Do not over-abstract for this small-scale project
- Keep it simple and practical
- Avoid unnecessary layers or indirection

---

# Coding Guidelines

- Prefer Go standard library whenever possible.
- Avoid premature use of interfaces.
- Functions should be small, explicit, and focused.
- Use clear and descriptive names.
- Always pass context.Context where applicable.
- Errors should be wrapped and returned, not swallowed.

This project is not intended to be used as a library.
godoc-style documentation is not required.

---

# Comments Policy

- Avoid excessive or redundant comments.
- Code should be self-explanatory through naming and structure.
- Do not explain what the code is doing if it is obvious from the code itself.

Allowed comments:
- Explaining why a decision was made
- Clarifying non-obvious constraints or trade-offs
- Linking to external documentation or specifications

Avoid comments such as:
- Line-by-line explanations
- Restating function or variable names
- Explaining obvious control flow

Design preference:
- Favor self-documenting code over comments.
- If a comment seems necessary, first consider whether the code can be made clearer instead.

---

# Scraping Policy

- Scraping targets are fixed and explicitly defined.
- URLs and parsing rules are implementation details.
- Expect target HTML structures to change over time.
- Handle scraping failures gracefully.

Do not:
- Automatically expand scraping targets
- Guess or crawl unrelated pages
- Introduce generic crawling logic

---

# Before Committing

Before Committing

Always run:

```
task tidy
task ci-check
```

## If ci-check fails

### 1. golangci-lint errors
You MUST fix lint issues using the auto-fix command:

```
task lint-fix
```

DO NOT manually edit any files before running `task lint-fix`.

### 2. goreg import organization errors
Fix import issues using the provided goreg task:

```
task goreg-fix
```

If needed, you may run per-file fixes:

```
goreg -w <filename>
```

Manual edits are allowed ONLY for the files reported by goreg.

DO NOT modify unrelated files.

### 3. other errors
You CAN edit files manually.

## Re-run verification
After fixing issues, run:

```
task ci-check
```

---

# Instructions for Claude

- Make the smallest reasonable change.
- Preserve existing behavior unless explicitly instructed otherwise.
- Do not add comments unless they provide meaningful context.
- If assumptions are unclear, ask before implementing.
- Briefly explain intent before making non-trivial changes.
- After editing or creating a `.go` file, run `goreg -w <file>` to organize imports.
