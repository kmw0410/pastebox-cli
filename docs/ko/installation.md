# 설치

GitHub Release는 Pastebox CLI용 Linux 패키지를 제공합니다.

| 배포판 | amd64 | arm64 |
|---|---|---|
| Debian / Ubuntu | `amd64.deb` | `arm64.deb` |
| Arch Linux 계열 | `x86_64.pkg.tar.zst` | `aarch64.pkg.tar.zst` |
| RHEL 계열 | `x86_64.rpm` | 제공하지 않음 |
| Void (glibc) | `x86_64.xbps` | `aarch64.xbps` |
| Void (musl) | `x86_64-musl.xbps` | `aarch64-musl.xbps` |

대상 시스템에 맞는 릴리스 패키지를 내려받은 뒤 배포판의 패키지 관리자로 설치합니다.

```bash
sudo apt install ./pastebox-cli_VERSION-1_amd64.deb
sudo pacman -U ./pastebox-cli-VERSION-1-x86_64.pkg.tar.zst
sudo dnf install ./pastebox-cli-VERSION-1.x86_64.rpm
```

`pb update`는 지원되는 최신 DEB 또는 RPM 패키지를 내려받고 검증한 후 설치합니다.
Arch Linux 계열에서는 `paru` 또는 `yay`로 AUR 패키지 업데이트를 확인하고
가능한 경우 설치합니다. AUR 패키지는 `paru -S pastebox-cli` 또는
`yay -S pastebox-cli`로 설치할 수 있습니다.

패키지 설치, 검증, 제거의 전체 절차는 [package_ko.md](../../package_ko.md)를
참조합니다.

## Nix

저장소에서 `x86_64-linux`와 `aarch64-linux`용 flake를 제공합니다.
Nix의 `nix-command`, `flakes` 실험적 기능을 활성화한 뒤 실행합니다.

```bash
nix run github:kmw0410/pastebox-cli -- version
nix profile add github:kmw0410/pastebox-cli#pastebox-cli
```

로컬 저장소에서는 `nix build .#pastebox-cli`로 빌드하고
`./result/bin/pb version`으로 실행합니다. 개발 중 Git 기반 flake를 사용하려면
새 Nix 파일이 Git 추적 대상이어야 합니다. `nix build path:.#pastebox-cli`는
추적되지 않은 파일도 포함합니다.
Nix는 Nix store에 설치하고 프로필을 통해 `pb` 명령을 제공합니다.
이 설치본은 `pb update` 대신 Nix로 관리합니다.

의존성과 버전 관리 방법은 [Nix 패키징](./packaging.md#nix)을 참조합니다.
