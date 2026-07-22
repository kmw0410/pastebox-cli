# Pastebox CLI
Pastebox 텍스트 업로드/원문 조회용 터미널 클라이언트

[English](./README.md) | Korean

패키지 문서: [설치 및 사용법](./package_ko.md)

### 기술 스택
| 레이어 | 스택 |
|--------|------|
| 언어 | Go 1.26.4 |
| 전송 | Go 표준 라이브러리 HTTP 클라이언트 |
| 패키징 | nFPM |
| 릴리스 | GitHub Actions + GitHub Releases |

### 디렉터리 구조
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

Debian, Arch, RPM 패키지 빌드 워크플로는 기존 릴리스 태그를 지정하여 각각
수동 실행할 수 있습니다. 자동 릴리스에서는 `cli-package-build.yml`이 세
워크플로를 호출하고 결과물을 합쳐 `release.yml`로 전달합니다.
`aur-publish.yml`은 릴리스 이후 독립적으로 게시하는 워크플로로 유지됩니다.

### AUR 패키징

저장소 루트의 `PKGBUILD`와 `.SRCINFO`는 소스 기반 `pastebox-cli` AUR
패키지를 정의합니다. 이 저장소 자체는 AUR 저장소가 아니며, AUR에
자동으로 게시하거나 푸시하지 않습니다.

새 릴리스를 준비할 때는 다음 순서를 따릅니다.

1. `PKGBUILD`의 `_tag`를 정확한 Git 릴리스 태그로 설정합니다.
2. `pkgver`는 태그의 `v`를 제거하고 `-`를 `.`으로 바꿔 설정합니다.
3. `_commit`을 태그가 가리키는 짧은 커밋 ID로 설정합니다.
4. 새 업스트림 버전에서 `pkgrel`을 `1`로 초기화합니다.
5. 소스 체크섬을 갱신하고 `.SRCINFO`를 다시 생성합니다.

   ```bash
   updpkgsums
   makepkg --printsrcinfo > .SRCINFO
   ```

`pb version`이 GitHub Release와 같은 태그를 표시하도록 `_tag`는 원래 태그를
유지하고, Arch 패키지 버전에는 정규화한 `pkgver`를 사용합니다.

Arch Linux 시스템에서 저장소 루트를 기준으로 검증합니다.

```bash
makepkg --verifysource
makepkg --cleanbuild
makepkg --printsrcinfo > .SRCINFO
namcap PKGBUILD
namcap pastebox-cli-*.pkg.tar.zst
./pkg/pastebox-cli/usr/bin/pb version
```

`namcap`은 선택 검증 의존성입니다. AUR에 게시할 때는 `PKGBUILD`와
`.SRCINFO`만 별도 AUR Git 저장소로 복사하고, 해당 저장소에 필요한
메인테이너 주석을 추가합니다.

### 어떻게 사용하나요?
1. GitHub Release에서 맞는 패키지를 내려받거나 `go build`로 직접 빌드합니다.
2. 저장소에 포함된 `config.json` 예시를 `~/.config/pastebox/config.json`으로 복사하거나, `pb`를 한 번 실행해 자동 생성합니다.
3. `pb config set server <URL>`로 서버를 설정한 뒤 `pb`로 업로드하고 `pb show`로 원문을 조회합니다.

### 명령어
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

명령별 사용법은 `pb show --help`, `pb clone --help`, `pb delete --help`, `pb manage --help`, `pb config --help` 또는 `pb update --help`로
확인할 수 있습니다. Arch Linux 계열에서 `pb update`는 최신 릴리스를 확인한
뒤 설치된 `paru` 또는 `yay`로 AUR 패키지를 업데이트합니다. Debian/Ubuntu
및 지원되는 RHEL/Fedora 계열에서는 최신 GitHub Release 패키지를 내려받아
검증한 후 설치합니다.
진행 중인 네트워크 요청은 `Ctrl-C`로 취소할 수 있으며, 전체 업로드 시간을
제한하지 않으면서 연결, TLS 핸드셰이크, 응답 헤더 대기 시간만 제한합니다.

### 기능
1. **스트리밍 업로드**: 파일명 보존 업로드와 stdin 파이프 입력을 모두 지원하며, 전체 입력을 메모리에 올리지 않습니다.

   ```bash
   pb server.log
   journalctl -u nginx | pb
   ```

2. **보관 정책 제어**: 영구, 일회성, 사용자 지정 만료 업로드를 지원합니다.

   ```bash
   pb --permanent config.yaml
   pb --once incident.txt
   pb --expires 12h build.log
   ```

3. **비밀번호 보호 Paste 조회**: 비밀번호를 화면에 표시하지 않는 프롬프트로 입력받아 `paste-password` 헤더로 전달합니다. 비밀번호는 셸 기록이나 프로세스 인자에 남지 않습니다. 업로드와 복제의 비밀번호 프롬프트에서 값을 입력하지 않고 Enter를 누르면 서버가 랜덤 비밀번호를 생성하고, 값을 입력하면 확인 후 해당 비밀번호를 적용합니다. 사용자 지정 프롬프트 비밀번호에는 업로드·복제 응답의 `password_protected`를 확인해 주는 서버 버전이 필요하며, 확인할 수 없으면 CLI는 안전하게 실패합니다.

   ```bash
   pb show --password AbC123
   ```

4. **스크립트 친화적 출력**: 공개 URL만 출력하거나 JSON 형식으로 받을 수 있습니다.

   ```bash
   pb --quiet server.log
   pb --json server.log
   ```

5. **Paste 복제**: 기존 Paste를 복제하면서 새 보존 정책, 코드 또는 프롬프트로 입력한 비밀번호를 지정할 수 있습니다. 보호된 원본 Paste의 비밀번호도 별도 프롬프트로 입력합니다.

   ```bash
   pb clone AbC123
   pb clone --source-password --expires 12h --password AbC123
   ```

6. **Paste 삭제**: 업로드 결과의 삭제 URL을 전달하거나, Paste 코드만 전달한 뒤 숨김 프롬프트에 삭제 토큰을 입력할 수 있습니다.

   ```bash
   pb delete 'https://paste.example.com/AbC123?delete=DELETE_TOKEN'
   pb delete AbC123
   ```

7. **Paste 관리**: 비공개 관리 URL을 사용해 메타데이터를 확인하고 라벨·보존 정책 변경, 비밀번호 보호 활성화·해제 또는 삭제를 할 수 있습니다. 코드만 전달하면 관리 토큰을 프롬프트로 입력받습니다.

   ```bash
   pb manage show 'https://paste.example.com/AbC123?manage=MANAGE_TOKEN'
   pb manage label AbC123 '운영 로그'
   pb manage policy AbC123 permanent
   pb manage password enable AbC123
   pb manage password disable AbC123
   pb manage delete AbC123
   ```

### 릴리스 패키지
GitHub Release에서 다음 Linux 패키지를 제공합니다.

| 배포판 | amd64 | arm64 |
|---|---|---|
| Debian / Ubuntu | `amd64.deb` | `arm64.deb` |
| Arch Linux 계열 | `x86_64.pkg.tar.zst` | `aarch64.pkg.tar.zst` |
| RHEL 계열 | `x86_64.rpm` | 제공하지 않음 |

설치와 설정의 자세한 내용은 [package_ko.md](./package_ko.md)를 참고하세요.
