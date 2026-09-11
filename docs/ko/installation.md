# 설치

GitHub Release는 Pastebox CLI용 Linux 패키지를 제공합니다.

| 배포판 | amd64 | arm64 |
|---|---|---|
| Debian / Ubuntu | `amd64.deb` | `arm64.deb` |
| Arch Linux 계열 | `x86_64.pkg.tar.zst` | `aarch64.pkg.tar.zst` |
| RHEL 계열 | `x86_64.rpm` | 제공하지 않음 |

대상 시스템에 맞는 릴리스 패키지를 내려받은 뒤 배포판의 패키지 관리자로 설치합니다.

```bash
sudo apt install ./pastebox-cli_VERSION-1_amd64.deb
sudo pacman -U ./pastebox-cli-VERSION-1-x86_64.pkg.tar.zst
sudo dnf install ./pastebox-cli-VERSION-1.x86_64.rpm
```

`pb update`는 지원되는 최신 DEB 또는 RPM 패키지를 내려받고 검증한 후 설치합니다.
AUR 게시가 일시 중단되어 Arch Linux 계열의 자동 업데이트는 현재 제공하지 않습니다.

패키지 설치, 검증, 제거의 전체 절차는 [package_ko.md](../../package_ko.md)를
참조합니다.
