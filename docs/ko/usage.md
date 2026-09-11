# 사용법

## 업로드

파일을 업로드하거나 표준 입력을 스트리밍합니다.

```bash
pb server.log
journalctl -u nginx | pb
```

보관 정책에는 `--permanent`, `--once`, `--expires 12h`를 사용합니다. 공개 URL만
출력하려면 `--quiet`를, 구조화된 출력에는 `--json`을 사용합니다.

## Paste 조회 및 관리

```bash
pb show AbC123
pb clone AbC123
pb delete AbC123
pb manage show AbC123
```

보호된 Paste는 `pb show --password AbC123`으로 조회합니다. 비밀번호는 프롬프트로
입력받아 HTTP 헤더로 전달하며 명령행 인자에는 포함하지 않습니다.

명령별 옵션은 `pb <command> --help`로 확인합니다.
