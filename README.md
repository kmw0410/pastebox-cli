# Pastebox CLI
Pastebox terminal client for text upload and raw retrieval

English | [Korean](./README_ko.md)

Packages: [installation and usage](./package.md)

### Tech stack
| Layer | Stack |
|--------|------|
| Language | Go 1.26.4 |
| Transport | Go standard library HTTP client |
| Packaging | nFPM |
| Release | GitHub Actions + GitHub Releases |

### Directory structure
```text
pastebox-cli/
├── .github/
│   └── workflows/
│       ├── arch-package-build.yml
│       ├── aur-publish.yml
│       ├── cli-package-build.yml
│       ├── deb-package-build.yml
│       ├── release-build.yml
│       ├── release.yml
│       └── rpm-package-build.yml
├── .gitignore
├── .SRCINFO
├── AGENTS.md
├── LICENSE
├── PKGBUILD
├── README.md
├── README_ko.md
├── config.go
├── config.json
├── config_test.go
├── get.go
├── get_test.go
├── go.mod
├── main.go
├── main_test.go
├── output.go
├── package.md
├── package_ko.md
├── packaging/
│   └── nfpm.yaml
├── update.go
├── update_test.go
├── upload.go
├── upload_test.go
└── workflow_test.go
```

The Debian, Arch, and RPM package build workflows can each be run manually for
an existing release tag. `cli-package-build.yml` calls all three workflows for
the automatic release flow and combines their artifacts for `release.yml`.
`aur-publish.yml` remains an independent post-release publishing workflow.

### AUR packaging

The repository-root `PKGBUILD` and `.SRCINFO` files prepare the source-based
`pastebox-cli` AUR package. This repository is not an AUR repository and does
not publish or push anything to the AUR by itself.

For a new release:

1. Set `_tag` in `PKGBUILD` to the exact Git release tag.
2. Set `pkgver` to the tag without `v`, replacing each `-` with `.`.
3. Set `_commit` to the short commit ID referenced by the tag.
4. Reset `pkgrel` to `1` for a new upstream version.
5. Update the source checksum and regenerate `.SRCINFO`:

   ```bash
   updpkgsums
   makepkg --printsrcinfo > .SRCINFO
   ```

Keep the original tag in `_tag` so `pb version` matches GitHub Releases while
the normalized `pkgver` remains valid for an Arch package version.

Validate from the repository root on an Arch Linux system:

```bash
makepkg --verifysource
makepkg --cleanbuild
makepkg --printsrcinfo > .SRCINFO
namcap PKGBUILD
namcap pastebox-cli-*.pkg.tar.zst
./pkg/pastebox-cli/usr/bin/pb version
```

`namcap` is optional. Before a future AUR submission, copy only `PKGBUILD` and
`.SRCINFO` into the separate AUR Git repository and add the maintainer comment
expected for that repository.

### How to use?
1. Download a package from the matching GitHub Release, or build the binary locally with `go build`.
2. Copy the bundled `config.json` example to `~/.config/pastebox/config.json`, or run `pb` once to create it automatically.
3. Run `pb config set server <URL>`, then use `pb` for uploads and `pb show` for raw retrieval.

### Commands
```text
pb [options] [file|-]
pb show [--password] <code|url>
pb clone [options] <code|url>
pb delete <code|delete-url>
pb manage <command> [arguments]
pb config show
pb config set server <URL>
pb config validate
pb update
pb version
```

Run `pb show --help`, `pb clone --help`, `pb delete --help`, `pb manage --help`, `pb config --help`, or `pb update --help` for
command-specific usage. On Arch Linux family systems, `pb update` checks the
latest release and updates the AUR package with an installed `paru` or `yay`.
On Debian/Ubuntu and supported RHEL/Fedora systems, it
downloads, verifies, and installs the matching latest GitHub Release package.
Active
network requests can be canceled with `Ctrl-C`; connection setup, TLS handshake,
and response-header waits are bounded without imposing a total upload timeout.

### Features
1. **Streaming uploads**: Upload a file with its original filename or pipe stdin without loading the full input into memory.

   ```bash
   pb server.log
   journalctl -u nginx | pb
   ```

2. **Retention control**: Use permanent, one-time, or custom-expiration uploads.

   ```bash
   pb --permanent config.yaml
   pb --once incident.txt
   pb --expires 12h build.log
   ```

3. **Protected paste retrieval**: Prompt for a hidden password and send it through the `paste-password` header. Passwords do not appear in shell history or process arguments. For upload and clone password prompts, pressing Enter without a value requests a random server-generated password; entering a value asks for confirmation and applies that custom password. Custom prompted passwords require a server version that confirms `password_protected`; the CLI fails closed when that confirmation is unavailable.

   ```bash
   pb show --password AbC123
   ```

4. **Script-friendly output**: Print only the public URL or emit JSON.

   ```bash
   pb --quiet server.log
   pb --json server.log
   ```

5. **Clone a paste**: Copy an existing paste while choosing a new retention policy, code, or prompted password. Protected source pastes prompt for their password separately.

   ```bash
   pb clone AbC123
   pb clone --source-password --expires 12h --password AbC123
   ```

6. **Delete a paste**: Pass the delete URL returned by an upload, or pass a code and enter its delete token at the hidden prompt.

   ```bash
   pb delete 'https://paste.example.com/AbC123?delete=DELETE_TOKEN'
   pb delete AbC123
   ```

7. **Manage a paste**: Inspect metadata, change its label or retention policy, or enable or disable password protection with the private manage URL. A code-only target prompts for the manage token. Use `pb delete` to delete a paste.

   ```bash
   pb manage show 'https://paste.example.com/AbC123?manage=MANAGE_TOKEN'
   pb manage label AbC123 'production log'
   pb manage policy AbC123 permanent
   pb manage password enable AbC123
   pb manage password disable AbC123
   ```

### Release packages
GitHub Releases publish these Linux package formats:

| Distribution | amd64 | arm64 |
|---|---|---|
| Debian / Ubuntu | `amd64.deb` | `arm64.deb` |
| Arch Linux family | `x86_64.pkg.tar.zst` | `aarch64.pkg.tar.zst` |
| RHEL family | `x86_64.rpm` | Not provided |

Detailed install and config instructions live in [package.md](./package.md).
