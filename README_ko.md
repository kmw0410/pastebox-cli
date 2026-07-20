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

Debian, Arch, RPM 패키지 빌드 워크플로는 기존 릴리스 태그를 지정하여 각각
수동 실행할 수 있습니다. 자동 릴리스에서는 `cli-package-build.yml`이 세
워크플로를 호출하고 결과물을 합쳐 `release.yml`로 전달합니다.
`aur-publish.yml`은 릴리스 이후 독립적으로 게시하는 워크플로로 유지됩니다.

### 어떻게 사용하나요?
1. GitHub Release에서 맞는 패키지를 내려받거나 `go build`로 직접 빌드합니다.
2. 저장소에 포함된 `config.json` 예시를 `~/.config/pastebox/config.json`으로 복사하거나, `pb`를 한 번 실행해 자동 생성합니다.
3. `pb config set server <URL>`로 서버를 설정한 뒤 `pb`로 업로드하고 `pb get`으로 원문을 조회합니다.

### 명령어
```text
pb [options] [file|-]
pb get [--password PASSWORD] <code|url>
pb config show
pb config set server <URL>
pb config validate
pb version
```

명령별 사용법은 `pb get --help` 또는 `pb config --help`로 확인할 수 있습니다.
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

3. **비밀번호 보호 Paste 조회**: 원문 조회 시 `paste-password` 헤더로 비밀번호를 전달합니다.

   ```bash
   pb get --password 'PASTE_PASSWORD' AbC123
   ```

4. **스크립트 친화적 출력**: 공개 URL만 출력하거나 JSON 형식으로 받을 수 있습니다.

   ```bash
   pb --quiet server.log
   pb --json server.log
   ```

### 릴리스 패키지
GitHub Release에서 다음 Linux 패키지를 제공합니다.

| 배포판 | amd64 | arm64 |
|---|---|---|
| Debian / Ubuntu | `amd64.deb` | `arm64.deb` |
| Arch Linux 계열 | `x86_64.pkg.tar.zst` | `aarch64.pkg.tar.zst` |
| RHEL 계열 | `x86_64.rpm` | 제공하지 않음 |

설치와 설정의 자세한 내용은 [package_ko.md](./package_ko.md)를 참고하세요.
