# Pastebox CLI 패키지

`pastebox-cli` 패키지는 터미널에서 Pastebox로 텍스트를 업로드하고 원문을 조회하는 `pb` 명령을 설치합니다.

## 지원 패키지

GitHub Release에서 다음 Linux 패키지를 제공합니다.

| 배포판 | amd64 | arm64 |
|---|---|---|
| Debian / Ubuntu | `amd64.deb` | `arm64.deb` |
| Arch Linux 계열 | `x86_64.pkg.tar.zst` | `aarch64.pkg.tar.zst` |
| RHEL 계열 | `x86_64.rpm` | 제공하지 않음 |

## 설치

사용 중인 시스템과 아키텍처에 맞는 파일을 GitHub Release에서 내려받으세요.

Debian 또는 Ubuntu:

```bash
sudo apt install ./pastebox-cli_VERSION-1_amd64.deb
```

Arch Linux, Manjaro 또는 EndeavourOS:

```bash
sudo pacman -U ./pastebox-cli-VERSION-1-x86_64.pkg.tar.zst
```

x86-64 RHEL, Rocky Linux, AlmaLinux 또는 Fedora:

```bash
sudo dnf install ./pastebox-cli-VERSION-1.x86_64.rpm
```

64비트 ARM 시스템에서는 Debian `arm64` 패키지 또는 Arch `aarch64` 패키지를 사용합니다.
RPM 패키지는 x86-64 시스템에만 제공합니다.

## 설정

CLI는 사용자별로 다음 설정 파일만 읽습니다.

```text
~/.config/pastebox/config.json
```

저장소에는 다음 예시 파일도 포함되어 있습니다.

```text
./config.json
```

이 파일을 복사해서 사용할 수도 있고, 입력 없이 `pb`를 한 번 실행해 같은 형태로 자동 생성할 수도 있습니다.

```bash
mkdir -p ~/.config/pastebox
cp ./config.json ~/.config/pastebox/config.json
```

```bash
pb
```

```text
created config: /home/user/.config/pastebox/config.json
Run pb config set server <URL> before using pb.
```

생성된 파일은 사용자만 읽고 쓸 수 있는 `0600` 권한이며 `server_url`은 빈 값입니다. 파일을 열지 않고 실제 Pastebox 서버 주소를 설정할 수 있습니다.

```bash
pb config set server https://paste.example.com
```

이 명령은 URL을 검사하고 정규화한 후 설정 파일을 원자적으로 갱신합니다. 저장되는 파일은 다음과 같습니다.

```json
{
  "server_url": "https://paste.example.com"
}
```

`https://example.com/pastebox`처럼 하위 경로에 설치된 서버도 지원합니다. `pb`를 다시 실행해도 기존 설정 파일은 덮어쓰지 않습니다.

현재 설정을 표시할 수 있습니다.

```bash
pb config show
```

업로드 전에 설정을 검사할 수 있습니다.

```bash
pb config validate
```

설정에 문제가 있으면 파일 경로와 잘못된 필드 또는 URL을 구체적으로 표시합니다. 확인할 수 있는 JSON 오류에는 줄과 열 번호도 포함합니다.

## 업로드

원래 파일명을 보존하여 파일을 업로드합니다.

```bash
pb server.log
```

파이프로 전달된 텍스트를 업로드합니다.

```bash
journalctl -u nginx | pb
printf 'hello\n' | pb
```

기본값은 Pastebox 서버의 일반 임시 보관 정책입니다. 다른 보관 정책과 업로드 옵션도 사용할 수 있습니다.

```bash
pb --permanent config.yaml
pb --once message.txt
pb --expires 12h build.log
pb --password secret.txt
pb --code deploy-log --label "운영 배포" server.log
```

`--permanent`, `--once`, `--expires`는 함께 사용할 수 없습니다.

업로드에 성공하면 기본 출력에 공개 URL과 서버가 반환한 만료시간, 생성된 비밀번호, 비공개 manage URL 및 삭제 URL이 표시됩니다.

```text
url: https://paste.example.com/AbC123
expires: 2026-08-16T10:00:00+09:00
password: GENERATED_PASSWORD
manage: https://paste.example.com/AbC123?manage=MANAGE_TOKEN
delete: https://paste.example.com/AbC123?delete=DELETE_TOKEN
```

스크립트에서는 공개 URL만 출력하거나 JSON 출력을 사용할 수 있습니다.

```bash
URL="$(pb --quiet server.log)"
pb --json server.log
```

`--quiet`과 `--json`은 함께 사용할 수 없습니다.

## 원문 조회

Paste 코드 또는 공개 URL로 원문을 가져옵니다.

```bash
pb get AbC123
pb get https://paste.example.com/AbC123
pb get AbC123 > restored.log
```

비밀번호로 보호된 Paste는 다음과 같이 조회합니다.

```bash
pb get --password 'PASTE_PASSWORD' AbC123
```

CLI는 `config.json`에 설정된 서버에 속한 전체 URL만 허용합니다. 이를 통해 다른 호스트로 Paste 비밀번호가 잘못 전송되는 일을 방지합니다.

## 업데이트

지원되는 최신 패키지를 설치하려면 업데이트 명령을 직접 실행합니다.

```bash
pb update
```

Debian과 Ubuntu에서는 최신 GitHub Release에서 시스템에 맞는 `amd64` 또는
`arm64` DEB를 선택해 임시 파일로 스트리밍하고, GitHub가 게시한 SHA-256
digest를 검증한 다음 `apt-get`으로 설치합니다. x86-64 RHEL 계열과
Fedora에서는 같은 방식으로 RPM을 검증하고 `dnf`로 설치합니다. 패키지
관리자가 `sudo`를 통해 관리자 권한을 요청할 수 있습니다.

Arch Linux 계열에서는 최신 GitHub Release를 확인하지만 해당 Release의
패키지를 내려받지는 않습니다. 새 릴리스가 있으면 `paru`를 우선 사용하고,
없으면 `yay`를 사용해 다음 AUR 업데이트를 실행합니다.

```bash
paru -S pastebox-cli
# paru가 없으면
yay -S pastebox-cli
```

두 AUR 도구가 모두 없으면 `paru` 또는 `yay`를 설치한 뒤 `pb update`를 다시
실행하라고 안내하며, 이 경우는 오류로 처리하지 않습니다.

이미 최신 버전이면 변경하지 않습니다. ARM 시스템에서는 RPM 자동 업데이트를
지원하지 않습니다.

## 명령 도움말과 요청 취소

설정 파일을 읽거나 서버에 접속하지 않고 명령별 사용법을 확인할 수 있습니다.

```bash
pb get --help
pb config --help
pb update --help
```

진행 중인 업로드 또는 조회 요청은 `Ctrl-C`로 취소할 수 있습니다. 연결, TLS
핸드셰이크, 응답 헤더 대기에는 제한 시간이 적용되지만, 큰 스트리밍 입력을
중단할 수 있는 전체 요청 시간 제한은 적용하지 않습니다.

## 종료 코드

| 코드 | 의미 |
|---|---|
| `0` | 성공 |
| `1` | 네트워크, 서버 또는 출력 오류 |
| `2` | 잘못된 인수, 입력 또는 설정 |

## 다운로드 검증

각 Release에는 `checksums.txt`가 포함됩니다.

```bash
sha256sum --check checksums.txt
```

목록에 포함된 모든 CLI 패키지를 검사하므로 체크섬 파일과 패키지 파일을 같은 디렉터리에 두어야 합니다.

## 제거

Debian 또는 Ubuntu:

```bash
sudo apt remove pastebox-cli
```

Arch Linux 계열:

```bash
sudo pacman -R pastebox-cli
```

RHEL 계열 또는 Fedora:

```bash
sudo dnf remove pastebox-cli
```

패키지를 제거해도 `~/.config/pastebox/config.json`은 삭제하지 않습니다.
