# presence

![demo](demo.gif)

A typing test that doesn't care about your WPM.

## Philosophy

There's no score. No leaderboard. No backspace.

You type. You make mistakes. You keep going.

Mistakes stay - what's typed is typed. Like words already said.

`--fleeting` dissolves the quote when you're done. Even the things you get right don't last.

## Run

```bash
go build -o presence && ./presence
```

## Options

```
--daily     same quote all day
--fleeting  dissolve the quote into dust after completion
--quotes    path to a custom quotes JSON file
--api       fetch a quote from a URL
--version   print version
```

Bring your own quote:

```json
[{"text": "The soul becomes dyed with the color of its thoughts.", "author": "Marcus Aurelius"}]
```

Or fetch one:

```bash
./presence --api https://stoic.tekloon.net/stoic-quote
```

## License

MIT
