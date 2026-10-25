# Pastebox CLI

Pastebox 텍스트 업로드와 원문 조회를 위한 터미널 클라이언트입니다.

[English](./README.md) | Korean

설치, 설정, 사용법, 셸 자동 완성, 패키징 안내는 [문서](./docs/README.md)를
참조하세요.

## 디렉터리 구조

```text
pastebox-cli/
├── .github/workflows/    # CI 및 릴리스 워크플로
├── docs/
│   ├── en/               # 영문 문서
│   ├── ko/               # 한국어 문서
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

`paru -S pastebox-cli` 또는 `yay -S pastebox-cli`로 AUR에서 설치합니다.
`pb update`는 설치된 AUR 도구로 패키지 업데이트를 확인하고 설치합니다.
