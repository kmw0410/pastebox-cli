# 설정

Pastebox CLI는 다음 파일에서 실행 설정을 읽습니다.

```text
~/.config/pastebox/config.json
```

입력 없이 `pb`를 한 번 실행하여 파일을 생성한 뒤 서버 URL을 설정합니다.

```bash
pb
pb config set server https://paste.example.com
```

URL에는 `http://` 또는 `https://`를 사용해야 합니다. 예를 들어
`https://example.com/pastebox`처럼 경로 하위에 배포한 서버도 지원합니다.

다음 명령으로 현재 설정을 확인하거나 유효성을 검사합니다.

```bash
pb config show
pb config validate
```
