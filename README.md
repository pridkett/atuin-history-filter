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

That will just show your history and have it return just the command from the output. Now, how to inject that into your overall prompt varies by shell. I use [fish shell](https://fishshell.com/). Here's what I have in `fish` shell:

```fish
function _fzf_atuin_search_history --description "Use atuin and fzf to search command history. Replace the command line with the selected command."

    set FZF_DEFAULT_OPTS '--color=bg+:#3c3836,bg:#32302f,spinner:#fb4934,hl:#928374,fg:#ebdbb2,header:#928374,info:#8ec07c,pointer:#fb4934,marker:#fb4934,fg+:#ebdbb2,prompt:#fb4934,hl+:#fb4934'

    # use the │  (vertical bar) as separator - note: this is not a pipe
    set -f new_command (
        atuin-history-filter --print0 --ansi --header | 
        fzf --read0 \
            --ansi \
            --scheme=history \
            --multi \
            --delimiter ║\
            --nth 3.. \
            --accept-nth 3.. \
            --prompt "History> " \
            --preview-window="bottom:3:wrap" \
            --preview="string replace --regex '^.*?║.*?║ ' '' -- {} | fish_indent --ansi" \
            --preview-label="command preview" \
            --header-lines=1 \
            --reverse \
            --highlight-line \
            --border \
            --bind "ctrl-r:transform:
              # Inspect current prompt and cycle to the next one
              if test \"\$FZF_PROMPT\" = 'History> '
                  printf 'reload(atuin-history-filter --print0 --ansi --header -c)+change-prompt(History (pwd)> )'
              else if test \"\$FZF_PROMPT\" = 'History (pwd)> '
		  printf 'reload(ATUIN_SESSION=$ATUIN_SESSION atuin-history-filter --print0 --ansi --header -s)+change-prompt(History (session)> )'
              else if test \"\$FZF_PROMPT\" = 'History (session)> '
		  printf 'reload(atuin-history-filter --print0 --ansi --header --hostname)+change-prompt(History (host)> )'
              else
                  printf 'reload(atuin-history-filter --print0 --ansi --header)+change-prompt(History> )'
              end
          " \
            --print0 |
        string split0
    )

    if test $status -eq 0
        commandline --replace -- $new_command
    end

    commandline --function repaint
end


if not type -q atuin; or not type -q fzf
    return
end
bind \cr _fzf_atuin_search_history
```

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
