# presence

A typing test that doesn't care about your WPM.

Just you quote, and the quiet hum of your terminal.

Type the words. Notice the mistakes. Keep going anyway. The obstacle is the way.

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
--api       fetch a quote from the Stoic Quote API
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

## Terminal support

Works on both light and dark terminals. Colors adapt automatically.

## License

MIT — be present with it however you like.
