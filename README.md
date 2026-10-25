# Pastebox CLI

Pastebox terminal client for text upload and raw retrieval.

English | [Korean](./README_ko.md)

Read the [documentation](./docs/README.md) for installation, configuration,
usage, shell completion, and packaging guidance.

## Directory structure

```text
pastebox-cli/
├── .github/workflows/    # CI and release workflows
├── docs/
│   ├── en/               # English documentation
│   ├── ko/               # Korean documentation
│   └── README.md
├── packaging/
│   └── nfpm.yaml
├── AGENTS.md
├── PKGBUILD
├── .SRCINFO
├── flake.nix
├── flake.lock
├── package.nix
├── go.mod
├── go.sum
├── main.go
├── README.md
├── README_ko.md
├── package.md
└── package_ko.md
```

## AUR

[pastebox-cli](https://aur.archlinux.org/packages/pastebox-cli)

Install from AUR with `paru -S pastebox-cli` or `yay -S pastebox-cli`.
Run `pb update` to check and install package updates with the available helper.
