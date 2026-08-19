# Bash completion for nejen.
# Completes subcommand words dynamically from `nejen --internal-complete`,
# which lists the dispatcher's command registry, so it can never drift from
# the actual set of commands.

_nejen() {
  local cur="${COMP_WORDS[COMP_CWORD]}"
  local cmds=$(nejen --internal-complete 2>/dev/null)
  
  COMPREPLY=($(compgen -W "$cmds help" -- "$cur"))
}

complete -F _nejen nejen
