# 패키징

## Nix

루트 `package.nix`는 `buildGo127Module`을 사용하며, `flake.nix`에서
`packages.<system>.pastebox-cli`와 기본 패키지·앱으로 제공합니다.
`flake.lock`으로 nixpkgs를 고정합니다. Nix Build 워크플로는 지원하는 두 Linux
아키텍처의 네이티브 러너에서 빌드·테스트하고 설치 후 `pb version`도 확인합니다.

flake는 소스 날짜 기반 개발 버전과 가능한 경우 Git 리비전을 사용합니다.
릴리스 빌드에서는 `package.nix`를 호출할 때 `version`에 원래 릴리스 태그,
`revision`에 릴리스 커밋을 전달할 수 있습니다. 두 값은 `pb version`에
주입됩니다. 일반 경로 소스에는 Git 리비전이 없어 `unknown`으로 표시됩니다.

Go 의존성을 변경하면 `package.nix`의 `vendorHash`도 다시 계산합니다.
잠시 `lib.fakeHash`로 설정하고 `nix build .#pastebox-cli`를 실행한 뒤,
Nix 해시 불일치 오류의 실제 해시로 교체합니다. 임시 해시는 커밋하지 않습니다.
nixpkgs는 `nix flake update nixpkgs`로 명시적으로 갱신하고 다음을 실행합니다.

```bash
nix flake check --no-update-lock-file --print-build-logs
nix build .#pastebox-cli --no-update-lock-file
nix run . --no-update-lock-file -- version
```

Nix는 자체 패키지 정의와 store 경로를 사용하며, 배포판 아카이브 패키지는
계속 nFPM으로 정의합니다. 이 워크플로는 nixpkgs 등록이나 바이너리 캐시
게시를 수행하지 않습니다.

## AUR

저장소 루트의 `PKGBUILD`와 `.SRCINFO`는 소스 기반 AUR 패키지를 정의합니다. 새
릴리스에서는 다음을 수행합니다.

1. `_tag`를 정확한 Git 릴리스 태그로 설정합니다.
2. 태그에서 `v`를 제거하고 `-`를 `.`으로 바꾼 값을 `pkgver`로 설정합니다.
3. 릴리스 커밋의 짧은 ID를 `_commit`으로 설정합니다.
4. `pkgrel`을 `1`로 초기화합니다.
5. 체크섬과 `.SRCINFO`를 갱신합니다.

```bash
updpkgsums
makepkg --printsrcinfo > .SRCINFO
```

Arch Linux 시스템에서 패키지를 검증한 뒤 `PKGBUILD`와 `.SRCINFO`만 별도의 AUR
저장소로 복사합니다.

```bash
makepkg --verifysource
makepkg --cleanbuild
namcap PKGBUILD
```

## 워크플로 검증

`quality-checks.yml`은 GitHub 설정 및 테스트 파일 변경 시 push와 pull request에서
actionlint와 Go 워크플로 검증 테스트를 실행합니다. 읽기 전용 권한을 사용하며
릴리스를 생성하지 않습니다. 기존 릴리스 워크플로의 메타데이터 제외 조건과
태그 생성 전 품질 검증은 유지합니다.
