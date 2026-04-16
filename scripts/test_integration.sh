#!/usr/bin/env bash
#
# Integration test for gpeace.
# Creates a real Git merge conflict and verifies the parser works correctly.
# Then provides instructions for manual interactive testing.
#
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
BINARY="$PROJECT_DIR/gpeace"
TMPDIR=""

cleanup() {
    if [ -n "$TMPDIR" ] && [ -d "$TMPDIR" ]; then
        rm -rf "$TMPDIR"
    fi
}
trap cleanup EXIT

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

pass() { echo -e "${GREEN}PASS${NC}: $1"; }
fail() { echo -e "${RED}FAIL${NC}: $1"; exit 1; }
info() { echo -e "${YELLOW}INFO${NC}: $1"; }

# ── Step 1: Build the binary ──────────────────────────────────────────────────
info "Building gpeace..."
cd "$PROJECT_DIR"
go build -o "$BINARY" . || fail "Build failed"
pass "Binary built at $BINARY"

# ── Step 2: Run unit tests ───────────────────────────────────────────────────
info "Running unit tests..."
go test ./... || fail "Unit tests failed"
pass "All unit tests passed"

# ── Step 3: Create a temp repo with merge conflicts ──────────────────────────
TMPDIR=$(mktemp -d)
info "Creating test repo in $TMPDIR"

cd "$TMPDIR"
git init -b main test-repo >/dev/null 2>&1
cd test-repo
git config user.email "test@gpeace.dev"
git config user.name "gpeace-test"
git config commit.gpgsign false
git config gpg.format openpgp

# Base file
cat > app.py << 'PYEOF'
def greet(name):
    return f"Hello, {name}!"

def calculate(a, b):
    return a + b

def main():
    print(greet("World"))
    print(calculate(1, 2))

if __name__ == "__main__":
    main()
PYEOF

cat > config.json << 'JSONEOF'
{
    "version": "1.0.0",
    "debug": false,
    "port": 8080,
    "host": "localhost"
}
JSONEOF

git add -A
git commit -m "initial commit" >/dev/null 2>&1
pass "Base commit created"

# Branch A changes
git checkout -b branch-a >/dev/null 2>&1
cat > app.py << 'PYEOF'
def greet(name):
    return f"Hi there, {name}! Welcome!"

def calculate(a, b):
    """Adds two numbers with validation."""
    if not isinstance(a, (int, float)) or not isinstance(b, (int, float)):
        raise TypeError("Arguments must be numbers")
    return a + b

def main():
    print(greet("World"))
    print(calculate(1, 2))

if __name__ == "__main__":
    main()
PYEOF

cat > config.json << 'JSONEOF'
{
    "version": "2.0.0",
    "debug": true,
    "port": 9090,
    "host": "localhost",
    "log_level": "debug"
}
JSONEOF

git add -A
git commit -m "branch-a: update greet, add validation to calculate" >/dev/null 2>&1
pass "Branch-a commit created"

# Branch B changes (from main)
git checkout main >/dev/null 2>&1
git checkout -b branch-b >/dev/null 2>&1
cat > app.py << 'PYEOF'
def greet(name):
    return f"Hey {name}, good to see you!"

def calculate(a, b):
    """Sums two values."""
    result = a + b
    return result

def main():
    print(greet("World"))
    print(calculate(1, 2))

if __name__ == "__main__":
    main()
PYEOF

cat > config.json << 'JSONEOF'
{
    "version": "1.5.0",
    "debug": false,
    "port": 3000,
    "host": "0.0.0.0",
    "timeout": 30
}
JSONEOF

git add -A
git commit -m "branch-b: different greet, different calculate" >/dev/null 2>&1
pass "Branch-b commit created"

# Merge to create conflicts
git checkout branch-a >/dev/null 2>&1
git merge branch-b 2>&1 || true

# ── Step 4: Verify conflicts exist ──────────────────────────────────────────
CONFLICTS=$(git diff --name-only --diff-filter=U)
if [ -z "$CONFLICTS" ]; then
    fail "No conflicts generated"
fi
pass "Merge conflicts created in: $(echo $CONFLICTS | tr '\n' ' ')"

# ── Step 5: Show conflict markers ───────────────────────────────────────────
info "Conflict markers in app.py:"
echo "---"
cat app.py
echo "---"

# ── Step 6: Verify conflict markers are well-formed ─────────────────────────
info "Verifying conflict markers..."
for file in $CONFLICTS; do
    OPENERS=$(grep -c '^<<<<<<<' "$file" || true)
    SEPS=$(grep -c '^=======$' "$file" || true)
    CLOSERS=$(grep -c '^>>>>>>>' "$file" || true)
    if [ "$OPENERS" -eq 0 ]; then
        fail "$file has no <<<<<<< markers"
    fi
    if [ "$OPENERS" != "$SEPS" ] || [ "$SEPS" != "$CLOSERS" ]; then
        fail "$file has mismatched markers: <<<=$OPENERS ===$SEPS >>>=$CLOSERS"
    fi
    pass "$file has $OPENERS well-formed conflict block(s)"
done

# ── Step 7: Print summary ───────────────────────────────────────────────────
echo ""
echo "========================================================"
echo " Integration Test Summary"
echo "========================================================"
echo ""
echo " All automated checks passed!"
echo ""
echo " For manual interactive testing:"
echo "   cd $TMPDIR/test-repo"
echo "   $BINARY"
echo ""
echo " This will launch the interactive TUI where you can"
echo " resolve the merge conflicts in app.py and config.json."
echo ""
echo " After resolving, verify with:"
echo "   git diff --name-only --diff-filter=U  # should be empty"
echo "   cat app.py                             # no conflict markers"
echo "   cat config.json                        # no conflict markers"
echo "========================================================"
