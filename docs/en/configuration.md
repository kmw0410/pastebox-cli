# Configuration

Pastebox CLI reads its runtime configuration from:

```text
~/.config/pastebox/config.json
```

Run `pb` without input once to create the file, then set the server URL:

```bash
pb
pb config set server https://paste.example.com
```

The URL must use `http://` or `https://`. Deployments below a path, such as
`https://example.com/pastebox`, are supported.

Inspect or validate the active configuration with:

```bash
pb config show
pb config validate
```
