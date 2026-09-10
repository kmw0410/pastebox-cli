package main

import (
	"fmt"
	"strings"
)

const zshCompletion = `#compdef pb

_pb() {
  local -a upload_options clone_options
  upload_options=(
    '--permanent[keep the paste permanently]'
    '--once[delete after the first successful view]'
    '--expires=[expire after a duration]:duration:'
    '--password[prompt for password protection]'
    '--code=[use a custom paste code]:code:'
    '--label=[attach a paste label]:label:'
    '--quiet[print only the public URL]'
    '--json[print the JSON response]'
  )
  clone_options=(
    '--source-password[prompt for the source paste password]'
    '--permanent[keep the cloned paste permanently]'
    '--once[delete after the first successful view]'
    '--expires=[expire after a duration]:duration:'
    '--password[prompt for password protection for the clone]'
    '--code=[use a custom code for the clone]:code:'
    '--quiet[print only the cloned paste URL]'
    '--json[print the JSON response]'
  )

  if (( CURRENT == 2 )); then
    _values 'command or upload option' \
      'show[retrieve raw paste content]' \
      'clone[clone a paste]' \
      'delete[delete a paste]' \
      'manage[manage a paste]' \
      'config[manage configuration]' \
      'update[install the latest supported package]' \
      'completion[print shell completion]' \
      'help[show usage]' \
      'version[show version]' \
      '--help[show usage]' \
      '--version[show version]' \
      $upload_options
    return
  fi

  case $words[2] in
    completion)
      _values 'shell' zsh bash fish
      ;;
    show)
      _arguments '--password[prompt for the paste password]' '*:paste code or URL:'
      ;;
    clone)
      _arguments $clone_options '*:paste code or URL:'
      ;;
    delete)
      _arguments '1:paste code or delete URL:'
      ;;
    manage)
      if (( CURRENT == 3 )); then
        _values 'command' show label policy password
      elif [[ $words[3] == password && CURRENT == 4 ]]; then
        _values 'action' enable disable
      elif [[ $words[3] == policy && CURRENT == 5 ]]; then
        _values 'policy' temporary permanent once duration
      fi
      ;;
    config)
      if (( CURRENT == 3 )); then
        _values 'command' show validate set
      elif [[ $words[3] == set && CURRENT == 4 ]]; then
        _values 'field' server
      fi
      ;;
    update)
      ;;
    *)
      _arguments $upload_options '1:file:_files'
      ;;
  esac
}

_pb "$@"
`

const bashCompletion = `_pb() {
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
`

const fishCompletion = `complete -c pb -f

complete -c pb -n '__fish_use_subcommand' -a show -d 'Retrieve raw paste content'
complete -c pb -n '__fish_use_subcommand' -a clone -d 'Clone a paste'
complete -c pb -n '__fish_use_subcommand' -a delete -d 'Delete a paste'
complete -c pb -n '__fish_use_subcommand' -a manage -d 'Manage a paste'
complete -c pb -n '__fish_use_subcommand' -a config -d 'Manage configuration'
complete -c pb -n '__fish_use_subcommand' -a update -d 'Install the latest supported package'
complete -c pb -n '__fish_use_subcommand' -a completion -d 'Print shell completion'
complete -c pb -n '__fish_use_subcommand' -a help -d 'Show usage'
complete -c pb -n '__fish_use_subcommand' -a version -d 'Show version'

complete -c pb -n '__fish_use_subcommand' -l permanent -d 'Keep the paste permanently'
complete -c pb -n '__fish_use_subcommand' -l once -d 'Delete after the first successful view'
complete -c pb -n '__fish_use_subcommand' -l expires -r -d 'Expire after a duration'
complete -c pb -n '__fish_use_subcommand' -l password -d 'Prompt for password protection'
complete -c pb -n '__fish_use_subcommand' -l code -r -d 'Use a custom paste code'
complete -c pb -n '__fish_use_subcommand' -l label -r -d 'Attach a paste label'
complete -c pb -n '__fish_use_subcommand' -l quiet -d 'Print only the public URL'
complete -c pb -n '__fish_use_subcommand' -l json -d 'Print the JSON response'

complete -c pb -n '__fish_seen_subcommand_from completion' -a 'zsh bash fish'
complete -c pb -n '__fish_seen_subcommand_from show' -l password -d 'Prompt for the paste password'
complete -c pb -n '__fish_seen_subcommand_from clone' -l source-password -d 'Prompt for the source paste password'
complete -c pb -n '__fish_seen_subcommand_from clone' -l permanent -l once -l password -l quiet -l json
complete -c pb -n '__fish_seen_subcommand_from clone' -l expires -r
complete -c pb -n '__fish_seen_subcommand_from clone' -l code -r
complete -c pb -n '__fish_seen_subcommand_from manage' -a 'show label policy password'
complete -c pb -n '__fish_seen_subcommand_from manage; and __fish_seen_subcommand_from password' -a 'enable disable'
complete -c pb -n '__fish_seen_subcommand_from manage; and __fish_seen_subcommand_from policy' -a 'temporary permanent once duration'
complete -c pb -n '__fish_seen_subcommand_from config' -a 'show set validate'
complete -c pb -n '__fish_seen_subcommand_from config; and __fish_seen_subcommand_from set' -a server
`

func (a application) runCompletion(args []string) int {
	if len(args) == 1 && (args[0] == "--help" || args[0] == "-h") {
		fmt.Fprint(a.stdout, completionUsageText)
		return 0
	}
	if len(args) != 1 {
		fmt.Fprint(a.stderr, completionUsageText)
		return 2
	}

	var script string
	switch strings.ToLower(args[0]) {
	case "zsh":
		script = zshCompletion
	case "bash":
		script = bashCompletion
	case "fish":
		script = fishCompletion
	default:
		fmt.Fprintf(a.stderr, "unsupported shell %q; expected zsh, bash, or fish\n", args[0])
		return 2
	}
	fmt.Fprint(a.stdout, script)
	return 0
}
