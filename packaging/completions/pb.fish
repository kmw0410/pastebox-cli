complete -c pb -f

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
complete -c pb -n '__fish_seen_subcommand_from manage; and __fish_seen_subcommand_from show' -l json -d 'Print paste metadata as JSON'
complete -c pb -n '__fish_seen_subcommand_from manage; and __fish_seen_subcommand_from password' -a 'enable disable'
complete -c pb -n '__fish_seen_subcommand_from manage; and __fish_seen_subcommand_from policy' -a 'temporary permanent once duration'
complete -c pb -n '__fish_seen_subcommand_from config' -a 'show set validate'
complete -c pb -n '__fish_seen_subcommand_from config; and __fish_seen_subcommand_from set' -a server
