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
        commandline --replace -- (string replace --regex '^ +' '' -- $new_command)
    end

    commandline --function repaint
end


if not type -q atuin; or not type -q fzf
    return
end
bind \cr _fzf_atuin_search_history
