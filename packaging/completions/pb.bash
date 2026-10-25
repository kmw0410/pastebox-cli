_pb() {
  local cur command
  cur=${COMP_WORDS[COMP_CWORD]}
  command=${COMP_WORDS[1]}

  if (( COMP_CWORD == 1 )); then
    if [[ $cur == -* ]]; then
      COMPREPLY=( $(compgen -W '--permanent --once --expires --password --code --label --quiet --json --help --version' -- "$cur") )
    else
      COMPREPLY=( $(compgen -W 'show clone delete manage config update completion help version' -- "$cur") )
    fi
    return
  fi

  case $command in
    completion)
      COMPREPLY=( $(compgen -W 'zsh bash fish' -- "$cur") )
      ;;
    show)
      COMPREPLY=( $(compgen -W '--password' -- "$cur") )
      ;;
    clone)
      COMPREPLY=( $(compgen -W '--source-password --permanent --once --expires --password --code --quiet --json' -- "$cur") )
      ;;
    manage)
      if (( COMP_CWORD == 2 )); then
        COMPREPLY=( $(compgen -W 'show label policy password' -- "$cur") )
      elif [[ ${COMP_WORDS[2]} == show ]]; then
        COMPREPLY=( $(compgen -W '--json' -- "$cur") )
      elif [[ ${COMP_WORDS[2]} == password && COMP_CWORD == 3 ]]; then
        COMPREPLY=( $(compgen -W 'enable disable' -- "$cur") )
      elif [[ ${COMP_WORDS[2]} == policy && COMP_CWORD == 4 ]]; then
        COMPREPLY=( $(compgen -W 'temporary permanent once duration' -- "$cur") )
      fi
      ;;
    config)
      if (( COMP_CWORD == 2 )); then
        COMPREPLY=( $(compgen -W 'show set validate' -- "$cur") )
      elif [[ ${COMP_WORDS[2]} == set && COMP_CWORD == 3 ]]; then
        COMPREPLY=( $(compgen -W 'server' -- "$cur") )
      fi
      ;;
    update|delete)
      ;;
    *)
      COMPREPLY=( $(compgen -W '--permanent --once --expires --password --code --label --quiet --json' -- "$cur") )
      ;;
  esac
}

complete -F _pb pb
