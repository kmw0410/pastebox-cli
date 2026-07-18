# AGENTS.md

## 1. Project Overview
Pastebox CLI is the standalone terminal client for uploading text to a Pastebox server and retrieving raw pastes.

Core behavior:
- The installed package name is `pastebox-cli`.
- The installed command name is `pb`.
- The CLI communicates only through the public HTTP API.
- Do not import or depend on server-side `internal` packages from the Pastebox repository.

The project targets small static binaries, simple Linux packaging, and script-friendly output.

## 2. Directory Structure
```text
pastebox-cli/
├── .github/
│   └── workflows/
│       ├── cli-package-build.yml
│       ├── release-build.yml
│       └── release.yml
├── LICENSE
├── README.md
├── README_ko.md
├── AGENTS.md
├── WORK.md
├── config.json
├── go.mod
├── main.go
├── config.go
├── upload.go
├── get.go
├── output.go
├── *_test.go
├── package.md
├── package_ko.md
└── packaging/
    └── nfpm.yaml
```

## 3. Architecture
- Entry point: `main.go`
  - Command dispatch, argument parsing, exit codes, version output.
- Config handling: `config.go`
  - Config path resolution, validation, bootstrap file creation, URL normalization.
- Upload flow: `upload.go`
  - Multipart file uploads, stdin streaming uploads, headers, error mapping.
- Retrieval flow: `get.go`
  - Raw paste download, password header handling, URL/code validation.
- Output formatting: `output.go`
  - Human-readable output, quiet output, JSON passthrough.
- Packaging: `packaging/nfpm.yaml`
  - Linux package metadata and installed file layout.
- CI/CD: `.github/workflows/*.yml`
  - Tag creation, package build, release asset publication.

## 4. Critical Rules
- When asked which files were modified, respond only with files changed in the most recent edit scope.
- When committing, exclude `WORK.md` and `AGENTS.md` unless the user explicitly asks to include them.
- When committing, add a `Co-authored-by: Codex <codex@openai.com>` trailer unless the user explicitly asks not to.
- If a change scope would be split into 2 or more commits, ask before pushing; otherwise commit only unless the user explicitly says to push.
- When modifying markdown files, include the filename in the commit subject or body so the changed docs are obvious.
- Keep commit messages themselves in English and follow the existing short conventional style.
- Before starting work, always consult `WORK.md` alongside `AGENTS.md` so recent task history is part of the working context.
- If `WORK.md` does not exist yet, create it before the first commit+push sequence that should be logged.
- After any commit+push sequence, record in `WORK.md` what work was done and what mostly changed, using the existing date-based log format, and include the commit ID plus commit message when available.
- If a follow-up fix is needed because of the agent's own mistake, record that in `WORK.md` as well, including what was wrong and how it was corrected.
- In `WORK.md`, every work-entry line under a date heading must start with `- ` so dated blocks are easy to scan.
- When a problem, regression, or unexpected behavior is reported or discovered, inspect the relevant git history before fixing it. Use commands such as `git log --follow -p -- <file>` and `git blame <file>` to identify when the affected code changed and what part changed.
- When fixing code after an error occurs, record the problematic code area, the root cause, and the fix in `AGENTS.md`, including a small relevant code snippet when useful.
- Before making a similar future change, review those recorded error-fix notes to avoid repeating the same mistake.

## 5. Configuration Rules
- Read configuration only from `~/.config/pastebox/config.json`.
- Do not add profiles, alternate config paths, `--server`, `PASTEBOX_URL`, or `XDG_CONFIG_HOME` support unless explicitly requested.
- The repository may include `config.json` as an example file, but runtime configuration is still loaded only from `~/.config/pastebox/config.json`.
- Support only this structure initially:
  ```json
  {
    "server_url": "https://paste.example.com"
  }
  ```
- Allow only `http://` and `https://` server URLs.
- Allow deployments below a URL path.
- Reject query strings, fragments, and embedded user credentials.
- Validation errors must identify the config path and the precise problem.
- When JSON parse errors provide line and column information, surface them.

## 6. CLI Behavior
- Installed command name: `pb`.
- Supported commands:
  - `pb [options] [file|-]`
  - `pb get [--password PASSWORD] <code|url>`
  - `pb config validate`
  - `pb version`
- Keep CLI argument/config errors distinct from network/server errors through documented nonzero exit codes.
- Exit codes must remain:
  - `0`: success
  - `1`: network, server, or output failure
  - `2`: invalid arguments, input, or configuration

Upload behavior:
- File uploads must use multipart requests so original filenames are preserved.
- Stdin uploads must stream raw request bodies and must not load full input into memory.
- Support temporary uploads by default plus permanent, once, custom expiration, generated password, custom code, and label options.
- `--permanent`, `--once`, and `--expires` must remain mutually exclusive.
- `--quiet` and `--json` must remain mutually exclusive.

Retrieval behavior:
- `pb get` must request raw paste content.
- Paste passwords must be sent through the `paste-password` header, not a query string.
- Full paste URLs are allowed only when they match the configured server.

Output behavior:
- Keep normal human-readable output.
- Keep public-URL-only quiet output.
- Keep JSON output.

## 7. Security Rules
- Never print secrets in diagnostic logs.
- Normal successful output may show the returned password, manage URL, and delete URL because those are one-time user-facing results.
- Do not accidentally echo paste passwords in validation or debug failures.
- Do not loosen same-server checks for password-protected `pb get` URL handling without explicit approval.

## 8. Packaging Rules
- Build static CLI binaries with `CGO_ENABLED=0`.
- Linux target architectures must include:
  - `linux/amd64`
  - `linux/arm64`
- Use `packaging/nfpm.yaml` for package generation.
- Install the binary at `/usr/bin/pb`.
- Package docs under `/usr/share/doc/pastebox-cli/`.
- Debian package architecture mapping:
  - `amd64` -> `amd64`
  - `arm64` -> `arm64`
- Arch package architecture mapping:
  - `amd64` -> `x86_64`
  - `arm64` -> `aarch64`
- Normalize date-based Git tags for package versions so tags such as `v26.07.17-1` remain upgradeable on Debian and Arch, while `pb version` keeps the original tag and commit ID.

## 9. GitHub Actions
Release workflow split:
- `release-build.yml`
  - Creates a new release tag on `main` push or manual dispatch.
- `cli-package-build.yml`
  - Builds binaries and packages for an existing release tag.
  - Verifies package installation in Debian and Arch containers.
  - Publishes package artifacts.
- `release.yml`
  - Creates or validates the GitHub Release using uploaded package artifacts.

Current expected flow:
```text
Release Build -> CLI Package Build -> Release
```

Workflow rules:
- Keep the package-build step independent from the server repository.
- Do not reintroduce assumptions about Docker image publishing here.
- Keep manual dispatch paths usable for rebuilding packages or recreating releases from existing artifacts.

## 10. Testing Rules
- Run `go test ./...` in this repository because it is a standalone module.
- Prefer local HTTP test servers for upload/get tests.
- Cover:
  - file and stdin streaming
  - config validation diagnostics
  - URL joining and normalization
  - output modes
  - password header handling
  - HTTP error mapping
  - exit codes
  - secret-safe failures

## 11. Documentation Rules
- Keep package installation and usage docs in `package.md` and `package_ko.md`.
- Keep top-level repository docs in `README.md` and `README_ko.md`.
- Keep English and Korean docs aligned when behavior changes.
- If config bootstrap behavior changes, update both README and package docs together.

## 12. Common Mistakes to Avoid
- Adding server-only dependencies or imports.
- Switching stdin uploads to full-buffer reads.
- Sending paste passwords through query strings.
- Letting `pb get` accept URLs from a different host than the configured server.
- Breaking exit-code separation between user errors and network/server failures.
- Forgetting to update both English and Korean docs.
- Reintroducing packaging paths that point back to the Pastebox server repository.

## 13. Detailed Work Log
- Each dated `WORK.md` block must use at least two lines.
  - Each work-entry line under the date heading must start with `- `.
  - Line 1 should summarize what changed.
  - Line 2 should spell out the concrete code/file or behavior changes.
  - If a commit+push happened, add the commit ID and commit message on the same dated block.
  - If the same calendar date already exists, append the new notes under that date instead of creating a duplicate date heading.

## 14. Error Fix Notes
Use this section to record recurring lessons from error-driven code fixes. Each note should include:
- Problematic code area: file/function or behavior that failed.
- Cause: why the error happened.
- Fix: how it was corrected, with a small code snippet when helpful.
- Prevention: what to check before making similar changes again.

No CLI-specific error-fix notes are recorded yet.

## 15. Commit Message Examples
Use short, conventional commit messages:
```text
feat: add config bootstrap file
feat: add CLI release workflows
fix: reject mismatched paste URLs
fix: keep stdin uploads streaming
docs: update package.md config examples
ci: add package verification jobs
```

## 16. Final Guidance
Keep Pastebox CLI:
- standalone
- HTTP-only
- stream-friendly
- packageable
- script-friendly
- minimal in dependencies

When editing, preserve:
- standalone Go module structure
- config validation strictness
- streaming upload behavior
- password header usage
- clear exit-code separation
- Linux package generation via nFPM
- split tag/build/release GitHub Actions flow
