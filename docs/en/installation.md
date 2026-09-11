# Installation

GitHub Releases provide Linux packages for Pastebox CLI.

| Distribution | amd64 | arm64 |
|---|---|---|
| Debian / Ubuntu | `amd64.deb` | `arm64.deb` |
| Arch Linux family | `x86_64.pkg.tar.zst` | `aarch64.pkg.tar.zst` |
| RHEL family | `x86_64.rpm` | Not provided |

Download the matching release package and install it with the native package
manager.

```bash
sudo apt install ./pastebox-cli_VERSION-1_amd64.deb
sudo pacman -U ./pastebox-cli-VERSION-1-x86_64.pkg.tar.zst
sudo dnf install ./pastebox-cli-VERSION-1.x86_64.rpm
```

Run `pb update` to download, verify, and install the latest supported DEB or
RPM package. Arch Linux family updates are temporarily unavailable because AUR
publication is paused.

For complete package installation, verification, and removal instructions, see
[package.md](../../package.md).
