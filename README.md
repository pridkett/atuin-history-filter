atuin-history-filter
====================

Patrick Wagstrom &lt;divine.baker9p@icloud.com&gt;

April 2025

This is a really simple little golang program that I created because I wanted a better way to manage my history in [atuin](https://atuin.sh/). In brief, I've gotten so used to piping my command history to [fzf](https://github.com/junegunn/fzf) that I don't want to move away from it, but I'd also like to spice it up a little bit. I want to know not only what commands, but I want to be able to see how many times I've run them, etc.

That's where this program comes in.

installation
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

	# use Gruvbox Dark color scheme in FZF
    set FZF_DEFAULT_OPTS '--color=bg+:#3c3836,bg:#32302f,spinner:#fb4934,hl:#928374,fg:#ebdbb2,header:#928374,info:#8ec07c,pointer:#fb4934,marker:#fb4934,fg+:#ebdbb2,prompt:#fb4934,hl+:#fb4934'

    # use the ║ (double vertical bar) as separator - this is better than
	# a comma or pipe because it far less common of a string
    set -f new_command (
        atuin-history-filter -print0 | fzf --read0 --scheme=history --multi --delimiter ║ --nth 3.. --accept-nth 3.. --prompt "History> " --preview-window="bottom:3:wrap" --preview="string replace --regex '^.*?║.*?║ ' '' -- {} | fish_indent --ansi" --print0 | string split0
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

License
-------

Copyright (c) 2025 Patrick Wagstrom

Licensed under terms of the MIT License
