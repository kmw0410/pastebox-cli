# Shell completion

DEB, Arch, RPM, and Void packages install Bash, Zsh, and Fish completion files
under `/usr/share/`. Nix installs the same files below its package's `share/`
directory. With the shell's completion support enabled, restart the shell to
load them; do not also append the scripts to startup files.

Print the completion script for the current shell and append it to that shell's
startup file. Restart the shell or source the file afterwards.

```bash
pb completion zsh >> ~/.zshrc
pb completion bash >> ~/.bashrc
pb completion fish >> ~/.config/fish/config.fish
```

## System-wide installation

Install the generated script in the completion directory for the target shell.
The directory may differ by distribution.

```bash
pb completion zsh | sudo tee /usr/share/zsh/site-functions/_pb > /dev/null
pb completion bash | sudo tee /usr/share/bash-completion/completions/pb > /dev/null
pb completion fish | sudo tee /usr/share/fish/vendor_completions.d/pb.fish > /dev/null
```

The Zsh filename is `_pb` because that is the completion function registered
for the `pb` command.
