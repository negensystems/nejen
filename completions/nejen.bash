# Bash completion for nejen.
#
# Names come from `nejen --internal-complete` -- the dispatcher's own
# registry, one command per line ("theme bg set") -- so they cannot drift.
# Completion walks those words one level at a time: `nejen theme <TAB>`
# offers bg, current, list, render, set.

_nejen() {
  local cur="${COMP_WORDS[COMP_CWORD]}"
  # Everything typed after "nejen", minus the word being completed.
  local typed="${COMP_WORDS[*]:1:COMP_CWORD-1}"
  local line rest
  local -a candidates=()

  while IFS= read -r line; do
    if [[ -z $typed ]]; then
      candidates+=("${line%% *}")
    elif [[ $line == "$typed "* ]]; then
      rest="${line#"$typed" }"
      candidates+=("${rest%% *}")
    fi
  done < <(nejen --internal-complete 2>/dev/null)

  if ((${#candidates[@]} == 0)); then
    # Past the last subcommand word. Of theme name / duration / file, a
    # path is the only one worth guessing at.
    COMPREPLY=($(compgen -f -- "$cur"))
    return
  fi

  # Commands sharing a prefix repeat the same next word. Offer it once.
  COMPREPLY=($(compgen -W "$(printf '%s\n' "${candidates[@]}" | sort -u)" -- "$cur"))
}

complete -F _nejen nejen
