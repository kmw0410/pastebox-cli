# Usage

## Upload

Upload a file or stream standard input:

```bash
pb server.log
journalctl -u nginx | pb
```

Choose a retention policy with `--permanent`, `--once`, or `--expires 12h`.
Use `--quiet` to print only the public URL, or `--json` for structured output.

## Retrieve and manage pastes

```bash
pb show AbC123
pb clone AbC123
pb delete AbC123
pb manage show AbC123
```

`pb show --password AbC123` prompts for a protected paste password. The value
is sent in an HTTP header and is not placed in command arguments.

Run `pb <command> --help` for command-specific options.
