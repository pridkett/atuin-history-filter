atuin-history-filter
====================

Patrick Wagstrom &lt;divine.baker9p@icloud.com&gt;

April 2025

This is a really simple little golang program that I created because I wanted a better way to manage my history in [atuin](https://atuin.sh/). In brief, I've gotten so used to piping my command history to [fzf](https://github.com/junegunn/fzf) that I don't want to move away from it, but I'd also like to spice it up a little bit. I want to know not only what commands, but I want to be able to see how many times I've run them, etc.

That's where this program comes in.

Installation
------------

This is a golang program and pretty simple one at that. It really only works on systems where you can find your atuin history in `~/.local/share/atuin/history.db`. It also relies on the schema of atuin not really changing.

```bash
go install
```

Usage
-----

In most cases this should be used in conjunction with `fzf` to display your command history. That would go something like:

```bash
./atuin-history-filter --print0 | fzf --read0 --delimiter ║  --nth 3.. --accept-nth 3..
```

That will just show your history and have it return just the command from the output. Now, how to inject that into your overall prompt varies by shell. I use [fish shell](https://fishshell.com/). See [atuin_fzf_history.fish](atuin_fzf_history.fish) for the way that I manage to do that. This preseves the whole `Ctrl-r` to change between all commands, pwd, session, and host filters.

Command Line Options
--------------------

- `--include-deleted`, `-d`: Include deleted commands in the results.
- `--reverse`, `-r`: Reverse the sort order (oldest first).
- `--print0`, `-0`: Use null character as record separator. You'll usually want to use this if you're piping to `fzf` where you'll use the `--read0` option.
- `--cwd <dir>`, `-c`: Limit search to a specific directory (defaults to current directory)
- `--session <session>`, `-s`: Limit search to a specific session (defaults to `$ATUIN_SESSION`).
- `--db <path>`: Path to the database file.
- `--fieldsep <sep>`, `-f`: Field separator for output (defaults to `║`).
- `--ansi`, `-a`: Enable ANSI color output.
- `--header`: Print header before the results.
- `--header-last`: Print header after the results.
- `--hostname <value>`: Filter by hostname. If no value is provided, uses the $ATUIN_HOST_NAME environment variable or the system hostname.

License
-------

Copyright (c) 2025 Patrick Wagstrom

Licensed under terms of the MIT License
