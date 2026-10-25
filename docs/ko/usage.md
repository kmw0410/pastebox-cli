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
pb manage show --json AbC123
```

보호된 Paste는 `pb show --password AbC123`으로 조회합니다. 비밀번호는 프롬프트로
입력받아 HTTP 헤더로 전달하며 명령행 인자에는 포함하지 않습니다.

명령별 옵션은 `pb <command> --help`로 확인합니다.

HTTP 오류는 상태 코드와 표준 설명으로 표시합니다. Paste 내용과 비밀정보가
진단 출력에 노출되지 않도록 서버 오류 본문과 원본 통신 오류는 출력하지 않습니다.

잘못된 Paste 대상은 비밀번호 프롬프트나 HTTP 요청 전에 종료 코드 `2`로
처리합니다. 네트워크·서버·출력 오류의 종료 코드는 `1`입니다.

`pb manage show --json <code|manage-url>`은 스크립트에서 사용할 Paste 메타데이터를
JSON으로 출력하며 비밀번호와 비공개 토큰은 포함하지 않습니다. 다른 관리 명령은
사람이 읽기 쉬운 형식으로 출력합니다.
