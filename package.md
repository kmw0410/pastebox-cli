# Pastebox CLI Packages

`pastebox-cli` installs the `pb` terminal client for uploading text to and retrieving raw text from a Pastebox server.

## Supported packages

GitHub Releases provide packages for these Linux architectures:

| Distribution | amd64 | arm64 |
|---|---|---|
| Debian / Ubuntu | `amd64.deb` | `arm64.deb` |
| Arch Linux family | `x86_64.pkg.tar.zst` | `aarch64.pkg.tar.zst` |
| RHEL family | `x86_64.rpm` | Not provided |

## Install

Download the package for your system from the matching GitHub Release.

Debian or Ubuntu:

```bash
sudo apt install ./pastebox-cli_VERSION-1_amd64.deb
```

Arch Linux, Manjaro, or EndeavourOS:

```bash
sudo pacman -U ./pastebox-cli-VERSION-1-x86_64.pkg.tar.zst
```

RHEL, Rocky Linux, AlmaLinux, or Fedora on x86-64:

```bash
sudo dnf install ./pastebox-cli-VERSION-1.x86_64.rpm
```

Use the `arm64` Debian package or `aarch64` Arch package on 64-bit ARM systems.
The RPM package is available only for x86-64 systems.

## Configure

The CLI reads only this per-user configuration file:

```text
~/.config/pastebox/config.json
```

The repository also includes this example file:

```text
./config.json
```

You can copy that file into place, or run `pb` without input once to create the same structure automatically:

```bash
mkdir -p ~/.config/pastebox
cp ./config.json ~/.config/pastebox/config.json
```

```bash
pb
```

```text
created config: /home/user/.config/pastebox/config.json
Run pb config set server <URL> before using pb.
```

The generated file has user-only `0600` permissions and an empty `server_url`. Set your Pastebox server without opening the file:

```bash
pb config set server https://paste.example.com
```

The command validates and normalizes the URL before atomically updating the configuration. The resulting file contains:

```json
{
  "server_url": "https://paste.example.com"
}
```

A server installed below a URL path is also supported, for example `https://example.com/pastebox`. Running `pb` again never overwrites an existing config file.

Show the active configuration:

```bash
pb config show
```

Validate the file before uploading:

```bash
pb config validate
```

The validator reports the config path and the specific malformed field or URL. JSON line and column numbers are included when available.

## Upload

Upload a file while preserving its original filename:

```bash
pb server.log
```

Upload piped text:

```bash
journalctl -u nginx | pb
printf 'hello\n' | pb
```

Pastebox uses its normal temporary retention policy by default. Other policies and upload options are available:

```bash
pb --permanent config.yaml
pb --once message.txt
pb --expires 12h build.log
pb --password secret.txt
pb --code deploy-log --label "Production deployment" server.log
```

`--permanent`, `--once`, and `--expires` cannot be combined.

The normal successful output contains the public URL and, when returned by the server, its expiration, generated password, private manage URL, and delete URL:

```text
url: https://paste.example.com/AbC123
expires: 2026-08-16T10:00:00+09:00
password: GENERATED_PASSWORD
manage: https://paste.example.com/AbC123?manage=MANAGE_TOKEN
delete: https://paste.example.com/AbC123?delete=DELETE_TOKEN
```

For scripts, print only the public URL or request JSON output:

```bash
URL="$(pb --quiet server.log)"
pb --json server.log
```

`--quiet` and `--json` cannot be combined.

## Retrieve raw text

Retrieve a paste by code or by a public URL:

```bash
pb show AbC123
pb show https://paste.example.com/AbC123
pb show AbC123 > restored.log
```

For a protected paste, provide its password through the request header with:

```bash
pb show --password 'PASTE_PASSWORD' AbC123
```

The CLI accepts full paste URLs only when they belong to the server configured in `config.json`. This prevents accidentally sending a paste password to another host.

## Update

Run the explicit update command to install the latest supported package:

```bash
pb update
```

On Debian and Ubuntu, the command selects the matching `amd64` or `arm64` DEB
from the latest GitHub Release, streams it to a temporary file, verifies the
SHA-256 digest published by GitHub, and installs it with `apt-get`. On x86-64
RHEL-family systems and Fedora, it performs the same process for the RPM and
installs it with `dnf`. The package manager may ask for administrator access
through `sudo`.

On Arch Linux family systems, the command checks the latest GitHub Release but
does not download a package from it. When a newer release exists, it uses
`paru` when available, or falls back to `yay`, to run the matching AUR update:

```bash
paru -S pastebox-cli
# or, when paru is unavailable
yay -S pastebox-cli
```

If neither AUR helper is installed, the command asks you to install `paru` or
`yay` and run `pb update` again; this guidance is not treated as an error.

An already current installation is left unchanged. Automatic RPM updates are
not available on ARM systems.

## Command help and cancellation

Show command-specific usage without reading the config file or contacting the server:

```bash
pb show --help
pb config --help
pb update --help
```

Press `Ctrl-C` to cancel an active upload or retrieval request. Connection setup,
TLS handshake, and response-header waits have bounded timeouts, while uploads do
not have a whole-request timeout that would interrupt large streaming inputs.

## Exit codes

| Code | Meaning |
|---|---|
| `0` | Success |
| `1` | Network, server, or output failure |
| `2` | Invalid arguments, input, or configuration |

## Verify downloads

Each release includes `checksums.txt`:

```bash
sha256sum --check checksums.txt
```

The command checks every CLI package listed in the file, so all listed package files must be present in the same directory.

## Remove

Debian or Ubuntu:

```bash
sudo apt remove pastebox-cli
```

Arch Linux family:

```bash
sudo pacman -R pastebox-cli
```

RHEL family or Fedora:

```bash
sudo dnf remove pastebox-cli
```

Package removal does not delete `~/.config/pastebox/config.json`.
