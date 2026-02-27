# presence

A typing test that doesn't care about your WPM.

Just you, a quote, and the quiet hum of your terminal.

Type the words. Notice the mistakes. Keep going anyway.

The obstacle is the way.

## Run

```bash
go run .
```

Or build it and sit with the binary for a while:

```bash
go build -o presence && ./presence
```

## Options

```
--random    pick a random quote instead of the daily one
--quotes    path to a custom quotes JSON file
--api       fetch a quote from an API endpoint (pass URL)
--version   print version and exit
```

Custom quotes file format:

```json
[
  {"text": "Your quote here.", "author": "You"}
]
```

## Controls

- **Type** — that's it, that's the app
- **Backspace** — for when the ego intervenes
- **Tab / Esc / Ctrl+C** — return to the world of distractions

## Start your day with a quote

Build and place the binary somewhere on your `$PATH`:

```bash
go build -o presence && mv presence /usr/local/bin/
```

Then add this to your `~/.zshrc` (or `~/.bashrc`):

```bash
# only run in interactive terminals, not inside editors/IDEs
if [[ $- == *i* && -z "$VSCODE_PID" && -z "$INTELLIJ_ENVIRONMENT_READER" ]]; then
  presence
fi
```

Every new terminal opens with a fresh quote to type through before you start your day.

## License

MIT — be present with it however you like.
