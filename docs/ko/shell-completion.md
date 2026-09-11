# 셸 자동 완성

사용 중인 셸의 자동 완성 스크립트를 출력하여 시작 파일에 추가합니다. 추가한 뒤
셸을 다시 시작하거나 해당 파일을 source 합니다.

```bash
pb completion zsh >> ~/.zshrc
pb completion bash >> ~/.bashrc
pb completion fish >> ~/.config/fish/config.fish
```

## 시스템 전역 설치

대상 셸의 자동 완성 디렉터리에 생성한 스크립트를 설치합니다. 디렉터리 경로는
배포판에 따라 달라질 수 있습니다.

```bash
pb completion zsh | sudo tee /usr/share/zsh/site-functions/_pb > /dev/null
pb completion bash | sudo tee /usr/share/bash-completion/completions/pb > /dev/null
pb completion fish | sudo tee /usr/share/fish/vendor_completions.d/pb.fish > /dev/null
```

Zsh의 파일명은 `pb` 명령에 등록된 자동 완성 함수 이름을 따라 `_pb`를
사용합니다.
