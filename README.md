# Pastebox CLI

Pastebox terminal client for text upload and raw retrieval.

English | [Korean](./README_ko.md)

Read the [documentation](./docs/README.md) for installation, configuration,
usage, shell completion, and packaging guidance.

## Directory structure

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
├── docs/
│   ├── en/
│   │   ├── packaging/
│   │   │   └── aur.md
│   │   ├── configuration.md
│   │   ├── installation.md
│   │   ├── shell-completion.md
│   │   └── usage.md
│   ├── ko/
│   │   ├── packaging/
│   │   │   └── aur.md
│   │   ├── configuration.md
│   │   ├── installation.md
│   │   ├── shell-completion.md
│   │   └── usage.md
│   └── README.md
├── packaging/
│   └── nfpm.yaml
├── AGENTS.md
├── LICENSE
├── PKGBUILD
├── README.md
├── README_ko.md
├── clone.go
├── clone_test.go
├── completion.go
├── config.go
├── config.json
├── config_test.go
├── delete.go
├── delete_test.go
├── get.go
├── get_test.go
├── go.mod
├── go.sum
├── main.go
├── main_test.go
├── manage.go
├── manage_test.go
├── output.go
├── package.md
├── package_ko.md
├── password_prompt.go
├── password_prompt_test.go
├── update.go
├── update_test.go
├── upload.go
├── upload_test.go
└── workflow_test.go
```

## [AUR](https://aur.archlinux.org/packages/pastebox-cli)

**AUR is temporarily unavailable.**
