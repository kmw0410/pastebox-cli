# AUR

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
