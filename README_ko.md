# Pastebox CLI

Pastebox 텍스트 업로드와 원문 조회를 위한 터미널 클라이언트입니다.

[English](./README.md) | Korean

설치, 설정, 사용법, 셸 자동 완성, 패키징 안내는 [문서](./docs/README.md)를
참조하세요.

## 디렉터리 구조

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

**AUR은 일시적으로 이용할 수 없습니다.**
