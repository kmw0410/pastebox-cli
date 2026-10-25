# Installation

GitHub Releases provide Linux packages for Pastebox CLI.

| Distribution | amd64 | arm64 |
|---|---|---|
| Debian / Ubuntu | `amd64.deb` | `arm64.deb` |
| Arch Linux family | `x86_64.pkg.tar.zst` | `aarch64.pkg.tar.zst` |
| RHEL family | `x86_64.rpm` | Not provided |
| Void (glibc) | `x86_64.xbps` | `aarch64.xbps` |
| Void (musl) | `x86_64-musl.xbps` | `aarch64-musl.xbps` |

Download the matching release package and install it with the native package
manager.

```bash
sudo apt install ./pastebox-cli_VERSION-1_amd64.deb
sudo pacman -U ./pastebox-cli-VERSION-1-x86_64.pkg.tar.zst
sudo dnf install ./pastebox-cli-VERSION-1.x86_64.rpm
```

Run `pb update` to download, verify, and install the latest supported DEB or
RPM package. On Arch Linux, it checks for a package update through `paru` or
`yay` and installs it when available. Install the AUR package with
`paru -S pastebox-cli` or `yay -S pastebox-cli`.

For complete package installation, verification, and removal instructions, see
[package.md](../../package.md).

## Nix

The repository provides a flake for `x86_64-linux` and `aarch64-linux`.
Enable the Nix `nix-command` and `flakes` experimental features, then run:

```bash
nix run github:kmw0410/pastebox-cli -- version
nix profile add github:kmw0410/pastebox-cli#pastebox-cli
```

From a local checkout, build with `nix build .#pastebox-cli` and run
`./result/bin/pb version`. When developing, Git-based flakes require new Nix
files to be tracked; `nix build path:.#pastebox-cli` also includes untracked files.
Nix installs the command in the Nix store and exposes `pb` through your profile.
Manage this installation with Nix rather than `pb update`.

See [Nix packaging](./packaging.md#nix) for dependency and version maintenance.
