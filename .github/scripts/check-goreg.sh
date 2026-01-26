#!/bin/bash
set -e

MODULE_PATH="github.com/Eagle-Konbu/shin-yokohama-event-notifier"
EXIT_CODE=0
FILES_WITH_ISSUES=()

echo "Checking import organization with goreg for changed files..."

# Get changed .go files in this PR compared to base branch
# Using git diff to get list of modified/added .go files
CHANGED_GO_FILES=$(git diff --name-only --diff-filter=AM origin/$GITHUB_BASE_REF...HEAD | grep '\.go$' || true)

if [ -z "$CHANGED_GO_FILES" ]; then
  echo "No .go files changed in this PR"
  exit 0
fi

echo "Changed .go files:"
echo "$CHANGED_GO_FILES" | sed 's/^/  - /'
echo ""

for file in $CHANGED_GO_FILES; do
  # Skip if file doesn't exist (edge case)
  if [ ! -f "$file" ]; then
    continue
  fi

  # Run goreg and capture output (without -w flag)
  GOREG_OUTPUT=$(goreg -l "$MODULE_PATH" "$file" 2>&1)

  # Compare with original file
  if ! echo "$GOREG_OUTPUT" | diff -u "$file" - > /dev/null 2>&1; then
    echo "❌ Import organization issue: $file"
    FILES_WITH_ISSUES+=("$file")
    EXIT_CODE=1

    # Show the diff for debugging
    echo "$GOREG_OUTPUT" | diff -u "$file" - || true
    echo ""
  fi
done

if [ $EXIT_CODE -ne 0 ]; then
  echo ""
  echo "================================================================"
  echo "The following files have import organization issues:"
  printf '  - %s\n' "${FILES_WITH_ISSUES[@]}"
  echo ""
  echo "To fix, run:"
  echo "  goreg -w -l $MODULE_PATH <file>"
  echo "================================================================"
  exit 1
fi

echo "✅ All changed Go files have properly organized imports"
