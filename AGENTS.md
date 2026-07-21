# AGENTS.md

## 1. Project Overview

Pastebox CLI is the standalone terminal client for uploading text to a Pastebox server and retrieving raw pastes.

- The installed package name MUST remain `pastebox-cli`.
- The installed command name MUST remain `pb`.
- The CLI MUST communicate with Pastebox only through its public HTTP API.
- The CLI MUST NOT import or depend on server-side `internal` packages from the Pastebox server repository.
- The project MUST remain a standalone Go module.
- Do not add a dependency unless the requested behavior requires it.
- Preserve small static binaries, Linux packaging, streaming input, and script-friendly output.

The terms in this document have these meanings:

- `MUST` and `MUST NOT` define mandatory requirements and prohibitions.
- `SHOULD` defines the default action; deviate only when the task provides a concrete reason.
- `MAY` defines an optional action.

## 2. Directory Structure

This tree lists the tracked files that define the current project. `WORK.md` is intentionally absent because it is an ignored local work log.

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
├── config.json
├── config.go
├── config_test.go
├── get.go
├── get_test.go
├── go.mod
├── main.go
├── main_test.go
├── output.go
├── package.md
├── package_ko.md
├── upload.go
├── upload_test.go
├── workflow_test.go
└── packaging/
    └── nfpm.yaml
```

## 3. Architecture

- `main.go` owns command dispatch, argument parsing, help and version output, request cancellation, HTTP client construction, and exit codes.
- `config.go` owns config bootstrap, loading, validation, atomic writes, and URL normalization.
- `upload.go` owns multipart file uploads, raw stdin streaming, upload headers, and HTTP error mapping.
- `get.go` owns raw paste retrieval, password headers, target validation, and redirect restrictions.
- `output.go` owns human-readable, quiet, and JSON upload output.
- `*_test.go` files test the standalone CLI with local HTTP servers and repository-file assertions.
- `packaging/nfpm.yaml` defines Debian, Arch, and RPM package metadata and installed paths.
- Root `PKGBUILD` and `.SRCINFO` define the source-based AUR package; both README variants document its maintenance workflow.
- `.github/workflows/` contains quality checks, tag creation, distribution package builds, GitHub Release creation, and AUR publication.

## 4. Required Work Sequence

1. Read `AGENTS.md`.
2. Read `WORK.md` if it exists. Continue if it does not exist.
3. Inspect the code, tests, configuration, workflows, or documentation that directly govern the request.
4. For a reported bug, regression, or unexpected behavior, inspect the relevant git history before editing.
5. Modify only the minimum files required by the user.
6. Run the validation required for the change type in Section 12.
7. Review the final diff for scope, correctness, secrets, and unintended user-file changes.
8. Do not commit or push unless the user requested that Git action.
9. Record local work in `WORK.md` when the work warrants a durable local note.
10. Report only files changed in the current edit scope and accurately list validations that ran.

Use `git log --follow -p -- <file>` and `git blame <file>` when investigating the history of a reported regression.

## 5. Configuration Rules

- Runtime configuration MUST be read only from `~/.config/pastebox/config.json`.
- The repository `config.json` MAY serve as an example, but runtime code MUST NOT load it as an alternate config.
- The supported configuration shape MUST remain:

  ```json
  {
    "server_url": "https://paste.example.com"
  }
  ```

- Do not add profiles unless the user explicitly requests profile support.
- Do not add alternate config paths unless the user explicitly requests a configuration-path feature.
- Do not add `--server`, `PASTEBOX_URL`, or `XDG_CONFIG_HOME` support unless the user explicitly requests that named interface.
- `server_url` MUST use `http://` or `https://`.
- `server_url` MUST support deployments below a URL path.
- `server_url` MUST reject query strings, fragments, and embedded user credentials.
- Config validation errors MUST identify the config path and the specific invalid field or URL property.
- JSON syntax and type errors MUST include line and column when the parser provides an offset.
- Preserve config bootstrap behavior and atomic `0600` config writes unless the user requests a config-write behavior change.

## 6. CLI Behavior

The primary usage list MUST match `main.go`:

```text
pb [options] [file|-]
pb get [--password PASSWORD] <code|url>
pb config show
pb config set server <URL>
pb config validate
pb version
```

- Preserve `pb help`, `pb --help`, `pb -h`, and `pb --version` aliases implemented by command dispatch.
- Preserve `pb get --help`, `pb get -h`, `pb config --help`, and `pb config -h`.
- Do not document an unimplemented command as supported.
- When a public command, option, argument order, or usage string changes, update the usage text and both language variants of the affected documentation in the same task.

Exit codes MUST remain:

- `0`: success.
- `1`: network, server, or output failure.
- `2`: invalid arguments, input, or configuration.

Do not merge argument or configuration failures into the network/server failure category.

### Upload

- File uploads MUST use multipart requests and preserve the original base filename.
- Stdin uploads MUST stream the raw request body without reading the full input into memory.
- Preserve temporary uploads by default.
- Preserve `--permanent`, `--once`, `--expires VALUE`, `--password`, `--code VALUE`, `--label VALUE`, `--quiet`, and `--json`.
- `--permanent`, `--once`, and `--expires` MUST remain mutually exclusive.
- `--quiet` and `--json` MUST remain mutually exclusive.

### Retrieval

- `pb get` MUST request raw paste content.
- Paste passwords MUST be sent in the `paste-password` HTTP header.
- Paste passwords MUST NOT be placed in a query string.
- A full paste URL MUST match the configured server.
- Redirects for `pb get` MUST retain the configured scheme, hostname, effective port, and base path.
- Redirect destinations MUST NOT contain user information.
- A rejected redirect MUST NOT receive the `paste-password` header.
- Do not weaken same-server or redirect checks unless the user explicitly requests a change to the retrieval trust boundary.

### Output

- Preserve normal human-readable output.
- Quiet mode MUST print only the public paste URL.
- Preserve JSON output.
- Successful upload output MAY show the password, manage URL, and delete URL returned for that upload.

## 7. Security Rules

- Diagnostic output MUST NOT contain paste passwords, authorization values, authentication headers, embedded URL credentials, tokens, or user paste contents.
- Errors MUST NOT echo the value of the `paste-password` header.
- Redirect errors MUST NOT expose user information from the destination URL.
- Successful user-facing upload output MAY contain the returned password, manage URL, and delete URL because they are results requested by the user.
- Tests, examples, `WORK.md`, and error-fix notes MUST use synthetic secrets and MUST NOT record real credentials or user data.

## 8. Packaging Rules

- Release binaries MUST be static and built with `CGO_ENABLED=0`.
- Package generation MUST use `packaging/nfpm.yaml`.
- Distribution packages MUST install the binary at `/usr/bin/pb`.
- nFPM packages MUST install documentation below `/usr/share/doc/pastebox-cli/`.
- Debian packages MUST support:
  - Go `amd64` -> Debian `amd64`.
  - Go `arm64` -> Debian `arm64`.
- Arch packages MUST support:
  - Go `amd64` -> Arch `x86_64`.
  - Go `arm64` -> Arch `aarch64`.
- RPM package builds MUST target `x86_64` only:
  - Go `amd64` -> RPM `x86_64`.
  - Go `arm64` -> unsupported.
- RPM `arm64` or `aarch64` packages MUST NOT be added unless the user explicitly requests RPM ARM support and the RPM workflow and package verification are updated in the same task.
- Preserve source-based AUR support for `x86_64` and `aarch64`.
- Date-based tags such as `v26.07.17-1` MUST be normalized into upgradeable distribution package versions.
- `pb version` MUST retain the original release tag and injected commit ID.

## 9. GitHub Actions Rules

- Keep workflow code independent from the Pastebox server repository.
- Do not add server Docker image publishing assumptions to this CLI repository.
- Preserve manual dispatch support for rebuilding packages or recreating releases from existing artifacts.
- Quality checks MUST complete successfully before a `main` push creates a release tag.
- Package workflows MUST preserve the architecture mappings in Section 8.
- Do not claim that a workflow or CI check ran when it was only reviewed locally.

## 10. Documentation Rules

- Keep repository overview and general usage in `README.md` and `README_ko.md`.
- Keep package installation and package-specific usage in `package.md` and `package_ko.md`.
- When public CLI behavior changes, update the English and Korean versions of every affected document in the same task.
- When config bootstrap behavior changes, update both README files and both package documentation files in the same task.
- Documentation MUST use command names, paths, filenames, package formats, and architectures that exist in the repository.
- For a documentation-only change, manually compare factual claims with the governing code, workflow, or packaging file.

## 11. Local Work Log

- `WORK.md` is an ignored local work log because `.gitignore` contains `WORK.md`.
- Read `WORK.md` at task start if it exists.
- A missing `WORK.md` MUST NOT fail or block the task.
- You MAY create `WORK.md` when a durable local work record is useful.
- `WORK.md` MUST NOT be staged or committed.
- Keep `WORK.md` entries local even when the related code is committed or pushed.
- You MAY record work whether or not a commit or push occurred.
- Do not describe unperformed work as completed.
- Do not describe uncommitted work as committed.
- If a commit exists, an entry MAY include its actual commit ID and commit message.
- If no commit exists, do not invent or estimate a commit ID.
- Append new entries below the existing heading for the same calendar date.
- Each work-entry line MUST start with `- `.
- Each dated work block MUST contain at least two work-entry lines: one summary and one concrete description.
- Record an agent-caused follow-up fix with the problem, cause, and correction.
- `WORK.md` MUST NOT contain passwords, tokens, authentication headers, private URLs, or user data.

## 12. Validation Rules

### Go code changes

- Apply `gofmt` to every changed `.go` file.
- Run `go test ./...`.
- Run `go build ./...`.

### CLI command or option changes

- Run `go test ./...`.
- Run `go build ./...`.
- Compare command dispatch and usage text with the documented commands and options.
- Update affected English and Korean documentation.

### Configuration changes

- Run `go test ./...`.
- Run `go build ./...`.
- Verify tests for the fixed config path, JSON diagnostics, URL validation, and bootstrap behavior.

### GitHub Actions or packaging changes

- Do not assume `actionlint` is installed.
- Manually inspect changed YAML and embedded shell commands.
- If the change can affect Go builds, run `go test ./...` and `go build ./...`.
- Compare build commands and package architecture mappings with the actual workflow and packaging files.
- Report CI validation as not run unless it actually ran on GitHub Actions.

### Documentation-only changes

- `go test ./...` and `go build ./...` are optional when no Go, workflow, or packaging behavior changed.
- If those commands are skipped, state that they were not run because the change was documentation-only.
- Manually verify documented commands, files, paths, package formats, and architectures against the current repository.

For upload and retrieval tests, use local `httptest` servers. Tests SHOULD cover streaming, config diagnostics, URL normalization, output modes, password headers, redirect restrictions, HTTP error mapping, exit codes, and secret-safe failures.

## 13. Git, Commit, and Push Rules

- If the user requests file changes but does not request a commit, the agent MUST NOT create a commit.
- Commit only when the user explicitly requests a commit.
- Push only when the user explicitly requests a push.
- When the user requests a commit and the work fits one commit, create that commit without asking for another approval.
- If the requested work requires two or more commits, tell the user before creating commits or pushing.
- Stage only files that belong to the requested commit.
- `WORK.md` MUST NOT be staged or committed.
- Exclude `AGENTS.md` from ordinary feature or fix commits unless the user explicitly requests its inclusion.
- Include `AGENTS.md` when the user explicitly requests changes to `AGENTS.md`.
- Existing user changes MUST NOT be overwritten or reverted.
- Do not run `git reset --hard`, `git clean -fd`, force push, or history-rewriting operations without explicit user approval for that named operation.
- Commit messages MUST be in English and follow the repository's short conventional style.
- When Markdown is a primary change, the commit subject or body MUST identify the changed documentation.
- Add `Co-authored-by: Codex <codex@openai.com>` to commits unless the user explicitly requests omission of that trailer.

Examples:

```text
feat: add config bootstrap file
fix: reject mismatched paste URLs
docs: update AGENTS.md repository rules
ci: add package verification jobs
```

## 14. Error Fix Notes

Keep general work and agent-caused follow-up details in `WORK.md`. Keep only repeatable failures here when future agents must check them before making the same class of change.

Each retained note MUST identify the problematic area, cause, fix, and prevention check. Never include real secrets.

### Source archive builds must disable automatic VCS stamping

- Problematic area: root `PKGBUILD` source archive builds.
- Cause: GitHub tag archives have no `.git` directory, so automatic Go VCS inspection failed with `error obtaining VCS status: exit status 128`.
- Fix: Disable automatic VCS stamping because release metadata is injected explicitly:

  ```bash
  go build -buildvcs=false \
    -ldflags "-X main.version=${_tag} -X main.commit=${_commit}"
  ```

- Prevention: Test extracted source archives, not only repository checkouts, and use `-buildvcs=false` when `-ldflags` supplies version metadata.

### nFPM environment assignments must use distinct source variables

- Problematic area: distribution package build steps in `.github/workflows/*-package-build.yml`.
- Cause: `PACKAGE_VERSION="$PACKAGE_VERSION" nfpm package` reused the destination environment name as its shell source, which triggered ShellCheck `SC2097` and `SC2098` and obscured expansion scope.
- Fix: Store the normalized value in a distinct shell variable:

  ```bash
  NORMALIZED_VERSION="${TAG#v}"
  PACKAGE_VERSION="$NORMALIZED_VERSION" nfpm package
  ```

- Prevention: Use distinct source and destination variable names. If `actionlint` is available, run it; otherwise manually inspect inline environment assignments and embedded shell.

### Redirect errors must not echo destination user information

- Problematic area: `get.go` redirect rejection errors for `pb get`.
- Cause: `http.Client.Do` wraps `CheckRedirect` errors in `url.Error`, whose text includes the rejected destination URL and can expose embedded credentials.
- Fix: Return a distinct application redirect error and unwrap it before formatting the request failure:

  ```go
  var redirectErr *getRedirectError
  if errors.As(err, &redirectErr) {
      return fmt.Errorf("request failed: %w", redirectErr)
  }
  ```

- Prevention: Test credential-bearing redirect URLs and assert that neither the destination request nor stderr receives a secret.

### Release path filters must preserve metadata exclusions

- Problematic area: `push.paths-ignore` and `pull_request.paths-ignore` in `.github/workflows/release-build.yml`.
- Cause: Adding quality checks replaced the existing metadata exclusions with a Markdown-only filter, so workflow, license, and ignore-file commits started release builds.
- Fix: Exclude Markdown, `.gitignore`, workflow files, and license files for both events; additionally exclude test-only Go changes from pushes while retaining pull-request validation for them.
- Prevention: Extend `workflow_test.go` whenever release path filters change and verify the intended push and pull-request exclusions independently.

## 15. Final Reporting

- Report only files modified in the current task.
- Report validation commands exactly as run and state whether each passed, failed, or was skipped.
- Distinguish local review from GitHub Actions execution.
- State whether a commit or push was performed.
