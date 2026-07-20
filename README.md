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
├── .SRCINFO
├── LICENSE
├── PKGBUILD
├── README.md
├── README_ko.md
├── config.json
├── package.md
├── package_ko.md
├── go.mod
├── main.go
├── config.go
├── config_test.go
├── upload.go
├── upload_test.go
├── get.go
├── get_test.go
├── output.go
└── packaging/
    ├── aur/
    │   └── README.md
    └── nfpm.yaml
```

The Debian, Arch, and RPM package build workflows can each be run manually for
an existing release tag. `cli-package-build.yml` calls all three workflows for
the automatic release flow and combines their artifacts for `release.yml`.
`aur-publish.yml` remains an independent post-release publishing workflow.

### How to use?
1. Download a package from the matching GitHub Release, or build the binary locally with `go build`.
2. Copy the bundled `config.json` example to `~/.config/pastebox/config.json`, or run `pb` once to create it automatically.
3. Run `pb config set server <URL>`, then use `pb` for uploads and `pb get` for raw retrieval.

### Commands
```text
pb [options] [file|-]
pb get [--password PASSWORD] <code|url>
pb config show
pb config set server <URL>
pb config validate
pb version
```

Run `pb get --help` or `pb config --help` for command-specific usage. Active
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

3. **Protected paste retrieval**: Fetch raw text while sending the password through the `paste-password` header.

   ```bash
   pb get --password 'PASTE_PASSWORD' AbC123
   ```

4. **Script-friendly output**: Print only the public URL or emit JSON.

   ```bash
   pb --quiet server.log
   pb --json server.log
   ```

### Release packages
GitHub Releases publish these Linux package formats:

| Distribution | amd64 | arm64 |
|---|---|---|
| Debian / Ubuntu | `amd64.deb` | `arm64.deb` |
| Arch Linux family | `x86_64.pkg.tar.zst` | `aarch64.pkg.tar.zst` |
| RHEL family | `x86_64.rpm` | Not provided |

Detailed install and config instructions live in [package.md](./package.md).
