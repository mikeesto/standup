# standup

A Go CLI that reads your recent git commits and uses an LLM (via OpenRouter) to generate a standup summary.

## Setup

Add your OpenRouter API key to `.zshrc` (or similar):

```bash
export OPENROUTER_MACBOOK_KEY=your_key_here
```

## Build & Run

```bash
# Build a binary and put it on your PATH
go build -o standup .
mv standup /usr/local/bin/

# Then from any git repo:
standup

# Include the last three rolling 24-hour periods:
standup --days 3
```

`--days` defaults to `1` and must be at least `1`.

## Output

```
Standup summary (last 24 hours)
--------------------------------
Raw commits:
- Add user authentication endpoint (a1b2c3d)
- Fix token expiry bug (e4f5g6h)
- Update README (i7j8k9l)

Summarising with AI...

Yesterday:
- Added user authentication endpoint with token-based login
- Fixed token expiry bug causing premature session logouts
- Updated README documentation
```
