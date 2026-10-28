package main

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestReleaseBuildRunsQualityChecksForPushAndPullRequest(t *testing.T) {
	workflow := readWorkflow(t, ".github/workflows/release-build.yml")
	for _, want := range []string{
		"push:\n    paths-ignore:\n      - \"**/*.md\"\n      - \".gitignore\"\n      - \".github/workflows/**\"\n      - \"**/LICENSE\"\n      - \"**/*_test.go\"",
		"pull_request:\n    paths-ignore:\n      - \"**/*.md\"\n      - \".gitignore\"\n      - \".github/workflows/**\"\n      - \"**/LICENSE\"",
		"go-version-file: go.mod",
		"cache: true",
		"unformatted=\"$(gofmt -l .)\"",
		"go vet ./...",
		"go test ./...",
		"go install golang.org/x/vuln/cmd/govulncheck@v1.9.0",
		"run: govulncheck ./...",
		"permissions:\n  contents: read",
	} {
		if !strings.Contains(workflow, want) {
			t.Errorf("release-build.yml does not contain %q", want)
		}
	}
}

func TestReleaseTagDependsOnQualityChecks(t *testing.T) {
	workflow := readWorkflow(t, ".github/workflows/release-build.yml")
	tagJob := strings.SplitN(workflow, "\n  tag:\n", 2)
	if len(tagJob) != 2 {
		t.Fatal("release-build.yml does not define the tag job")
	}
	if !strings.Contains(tagJob[1], "needs: test") {
		t.Fatal("release tag job does not depend on the test job")
	}
	if !strings.Contains(tagJob[1], "github.ref == 'refs/heads/main'") {
		t.Fatal("release tag job is not restricted to main pushes")
	}
	if !strings.Contains(tagJob[1], `BASE_TAG="v$(TZ=Asia/Seoul date +'%y.%m.%d')"`) {
		t.Fatal("release tag date is not generated in Asia/Seoul")
	}
	if strings.Contains(tagJob[1], "date -u") {
		t.Fatal("release tag date must not be generated in UTC")
	}
}

func TestAutomatedReleaseWorkflowsHonorSkipActionsMarker(t *testing.T) {
	for _, path := range []string{
		".github/workflows/release-build.yml",
		".github/workflows/cli-package-build.yml",
		".github/workflows/release.yml",
	} {
		workflow := readWorkflow(t, path)
		if !strings.Contains(workflow, "[skip ci]") {
			t.Errorf("%s does not honor the [skip ci] commit marker", path)
		}
	}
}

func TestAutomatedReleaseWorkflowsSkipDependabotCommits(t *testing.T) {
	for _, path := range []string{
		".github/workflows/release-build.yml",
		".github/workflows/cli-package-build.yml",
		".github/workflows/release.yml",
		".github/workflows/package-metadata-update.yml",
	} {
		workflow := readWorkflow(t, path)
		if !strings.Contains(workflow, "dependabot[bot]") {
			t.Errorf("%s does not skip Dependabot commits", path)
		}
	}
}

func TestAURPublishWorkflowRequiresSuccessfulReleaseOrManualDispatch(t *testing.T) {
	workflow := readWorkflow(t, ".github/workflows/aur-publish.yml")
	for _, want := range []string{"workflow_dispatch:", "- Release", "- main", "github.event_name == 'workflow_dispatch' ||", "github.event.workflow_run.conclusion == 'success'", "[skip ci]", "dependabot[bot]", "INPUT_TAG: ${{ inputs.tag }}", `TAG="$INPUT_TAG"`} {
		if !strings.Contains(workflow, want) {
			t.Errorf("AUR workflow is missing %q", want)
		}
	}
	if strings.Contains(workflow, "if: ${{ false }}") {
		t.Fatal("AUR publishing is still disabled")
	}
}

func readWorkflow(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestRPMVerificationInstallsDependenciesBeforeLocalPackage(t *testing.T) {
	workflow := readWorkflow(t, ".github/workflows/rpm-package-build.yml")
	for _, want := range []string{
		"rockylinux:9-minimal",
		"rpm -qp --requires /dist/${RPM_FILE} | grep -Fx ca-certificates",
		"microdnf install -y ca-certificates && rpm --install /dist/${RPM_FILE}",
		"test -s /etc/pki/tls/certs/ca-bundle.crt",
	} {
		if !strings.Contains(workflow, want) {
			t.Errorf("RPM verification is missing %q", want)
		}
	}
	if strings.Contains(workflow, "microdnf install -y /dist/") {
		t.Error("Rocky Linux 9 microdnf cannot install the local RPM file")
	}
	if strings.Contains(workflow, "--nodeps") {
		t.Error("RPM verification must enforce runtime dependencies")
	}
}

func TestWorkflowChangesAreCheckedWithoutCreatingReleases(t *testing.T) {
	workflow := readWorkflow(t, ".github/workflows/quality-checks.yml")
	for _, event := range []string{"push", "pull_request"} {
		if !strings.Contains(workflow, event+":\n    paths:\n      - \".github/**\"\n      - \"**/*_test.go\"") {
			t.Errorf("%s changes are not validated", event)
		}
	}
	for _, want := range []string{"workflow_dispatch:", "contents: read", "go-version-file: go.mod", "go vet ./...", "go test ./...", "actionlint@v1.7.12", "run: actionlint"} {
		if !strings.Contains(workflow, want) {
			t.Errorf("quality workflow is missing %q", want)
		}
	}
	for _, forbidden := range []string{"contents: write", "git tag", "git push", "gh release"} {
		if strings.Contains(workflow, forbidden) {
			t.Errorf("quality workflow contains release operation %q", forbidden)
		}
	}
}

func workflowStepScript(t *testing.T, path, name string) string {
	t.Helper()
	workflow := readWorkflow(t, path)
	_, step, ok := strings.Cut(workflow, "      - name: "+name+"\n")
	if !ok {
		t.Fatalf("missing step %s", name)
	}
	step = strings.SplitN(step, "\n      - name:", 2)[0]
	_, script, ok := strings.Cut(step, "        run: |\n")
	if !ok {
		t.Fatalf("missing script for %s", name)
	}
	var lines []string
	for _, line := range strings.Split(script, "\n") {
		if line != "" && !strings.HasPrefix(line, "          ") {
			break
		}
		lines = append(lines, strings.TrimPrefix(line, "          "))
	}
	return strings.Join(lines, "\n")
}

func runWorkflowScript(t *testing.T, script string, env map[string]string, mocks map[string]string) (string, error) {
	t.Helper()
	bash, err := exec.LookPath("bash")
	if err != nil {
		t.Skip("bash is required for workflow shell tests")
	}
	dir := t.TempDir()
	for name, body := range mocks {
		if err := os.WriteFile(filepath.Join(dir, name), []byte("#!/bin/sh\n"+body), 0700); err != nil {
			t.Fatal(err)
		}
	}
	cmd := exec.Command(bash, "--noprofile", "--norc", "-e", "-o", "pipefail", "-c", script)
	cmd.Env = append(os.Environ(), "PATH="+dir+string(os.PathListSeparator)+os.Getenv("PATH"), "GITHUB_OUTPUT="+filepath.Join(dir, "output"))
	for k, v := range env {
		cmd.Env = append(cmd.Env, k+"="+v)
	}
	output, err := cmd.CombinedOutput()
	return string(output), err
}

func TestExternalActionsArePinned(t *testing.T) {
	paths, err := filepath.Glob(".github/workflows/*.yml")
	if err != nil {
		t.Fatal(err)
	}
	pinned := regexp.MustCompile(`^[\w/-]+@[0-9a-f]{40}$`)
	for _, path := range paths {
		for _, line := range strings.Split(readWorkflow(t, path), "\n") {
			if !strings.HasPrefix(strings.TrimSpace(line), "uses: ") {
				continue
			}
			ref := strings.Fields(strings.TrimSpace(line))[1]
			if !strings.HasPrefix(ref, "./") && !pinned.MatchString(ref) {
				t.Errorf("%s: unpinned action %s", path, ref)
			}
		}
	}
}

func TestPackageTagInputCannotExecuteShell(t *testing.T) {
	script := workflowStepScript(t, ".github/workflows/cli-package-build.yml", "Resolve release tag")
	if strings.Contains(script, "${{") {
		t.Fatal("expressions must be passed through env")
	}
	marker := filepath.Join(t.TempDir(), "injected")
	output, err := runWorkflowScript(t, script, map[string]string{
		"EVENT_NAME": "workflow_dispatch", "INPUT_TAG": `v1$(touch "` + marker + `")`,
	}, map[string]string{"git": "exit 0\n"})
	if err == nil {
		t.Fatalf("unsafe tag accepted: %s", output)
	}
	if _, err := os.Stat(marker); !os.IsNotExist(err) {
		t.Fatal("tag input executed shell code")
	}
}

func TestReleaseRejectsUntrustedWorkflowRuns(t *testing.T) {
	if _, err := exec.LookPath("jq"); err != nil {
		t.Skip("jq is required for workflow validation tests")
	}
	script := workflowStepScript(t, ".github/workflows/release.yml", "Validate package workflow run")
	for _, name := range []string{"valid automatic", "valid manual", "wrong workflow", "wrong path", "fork", "wrong repo", "branch", "failed", "running", "pull request", "bad sha", "wrong id", "invalid input"} {
		t.Run(name, func(t *testing.T) {
			fixture := map[string]any{"id": 123, "workflow_id": 42, "path": ".github/workflows/cli-package-build.yml", "repository": map[string]string{"full_name": "example/cli"}, "head_repository": map[string]string{"full_name": "example/cli"}, "head_branch": "main", "status": "completed", "conclusion": "success", "event": "workflow_run", "head_sha": strings.Repeat("a", 40)}
			runID := "123"
			switch name {
			case "valid manual":
				fixture["event"] = "workflow_dispatch"
			case "wrong workflow":
				fixture["workflow_id"] = 43
			case "wrong path":
				fixture["path"] = ".github/workflows/other.yml"
			case "fork":
				fixture["head_repository"] = map[string]string{"full_name": "other/cli"}
			case "wrong repo":
				fixture["repository"] = map[string]string{"full_name": "other/cli"}
			case "branch":
				fixture["head_branch"] = "feature"
			case "failed":
				fixture["conclusion"] = "failure"
			case "running":
				fixture["status"] = "in_progress"
			case "pull request":
				fixture["event"] = "pull_request"
			case "bad sha":
				fixture["head_sha"] = "invalid"
			case "wrong id":
				fixture["id"] = 124
			case "invalid input":
				runID = "123/../../other"
			}
			data, err := json.Marshal(fixture)
			if err != nil {
				t.Fatal(err)
			}
			output, err := runWorkflowScript(t, script, map[string]string{"RUN_ID": runID, "DEFAULT_BRANCH": "main", "GITHUB_REPOSITORY": "example/cli", "RUN_FIXTURE": string(data)}, map[string]string{"gh": `case "$2" in
*/actions/workflows/cli-package-build.yml) echo 42 ;;
*/actions/runs/123) printf '%s\n' "$RUN_FIXTURE" ;;
*) exit 1 ;;
esac
`})
			valid := strings.HasPrefix(name, "valid ")
			if (err == nil) != valid {
				t.Fatalf("unexpected result: %v: %s", err, output)
			}
		})
	}
}

func TestReleaseArtifactMustMatchSourceCommit(t *testing.T) {
	script := workflowStepScript(t, ".github/workflows/release.yml", "Resolve release tag from package artifact")
	sha := strings.Repeat("a", 40)
	other := strings.Repeat("b", 40)
	for _, tc := range []struct {
		name, source, head, event string
		valid                     bool
	}{
		{"automatic", sha, sha, "workflow_run", true},
		{"manual old tag", sha, other, "workflow_dispatch", true},
		{"legacy matching", "", sha, "workflow_dispatch", true},
		{"legacy mismatching", "", other, "workflow_dispatch", false},
		{"moved tag", other, sha, "workflow_dispatch", false},
		{"invalid metadata", "invalid", sha, "workflow_dispatch", false},
		{"automatic mismatching head", sha, other, "workflow_run", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			pkg := filepath.Join(dir, "pastebox-cli-packages")
			if err := os.Mkdir(pkg, 0700); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(pkg, "release-tag.txt"), []byte("v26.10.28\n"), 0600); err != nil {
				t.Fatal(err)
			}
			if tc.source != "" {
				if err := os.WriteFile(filepath.Join(pkg, "release-commit.txt"), []byte(tc.source+"\n"), 0600); err != nil {
					t.Fatal(err)
				}
			}
			output, err := runWorkflowScript(t, script, map[string]string{"RUNNER_TEMP": dir, "RUN_HEAD_SHA": tc.head, "RUN_EVENT": tc.event, "TAG_COMMIT_FIXTURE": sha}, map[string]string{"git": `if [ "$1" = rev-parse ]; then printf '%s\n' "$TAG_COMMIT_FIXTURE"; fi
`})
			if (err == nil) != tc.valid {
				t.Fatalf("unexpected result: %v: %s", err, output)
			}
		})
	}
}

func TestPackageBuildersRejectMovedTags(t *testing.T) {
	for _, name := range []string{"deb", "arch", "rpm", "void"} {
		t.Run(name, func(t *testing.T) {
			script := workflowStepScript(t, ".github/workflows/"+name+"-package-build.yml", "Validate and checkout release tag")
			for _, expected := range []string{"", strings.Repeat("a", 40), strings.Repeat("b", 40)} {
				output, err := runWorkflowScript(t, script, map[string]string{"TAG": "v26.10.28", "EXPECTED_COMMIT": expected}, map[string]string{"git": "if [ \"$1\" = rev-parse ]; then echo " + strings.Repeat("a", 40) + "; fi\n"})
				if (err == nil) != (expected != strings.Repeat("b", 40)) {
					t.Fatalf("unexpected result: %v: %s", err, output)
				}
			}
		})
	}
}

func TestAURRejectsUnverifiedHostKeys(t *testing.T) {
	script := workflowStepScript(t, ".github/workflows/aur-publish.yml", "Configure AUR SSH access")
	for _, fingerprint := range []string{"SHA256:RFzBCUItH9LZS0cKB5UE6ceAYhBD5C8GeOBip8Z11+4", "SHA256:synthetic-wrong-key", ""} {
		dir := t.TempDir()
		output, err := runWorkflowScript(t, script, map[string]string{"HOME": dir, "RUNNER_TEMP": dir, "AUR_USERNAME": "synthetic", "AUR_EMAIL": "test@example.invalid", "AUR_SSH_PRIVATE_KEY": "synthetic-test-only", "FINGERPRINT_FIXTURE": fingerprint}, map[string]string{"ssh-keyscan": "echo synthetic-host-key\n", "ssh-keygen": `if [ -z "$FINGERPRINT_FIXTURE" ]; then exit 1; fi
printf '256 %s synthetic (ED25519)\n' "$FINGERPRINT_FIXTURE"
`})
		valid := strings.Contains(fingerprint, "RFzBCUIt")
		if (err == nil) != valid {
			t.Fatalf("unexpected result: %v: %s", err, output)
		}
		_, statErr := os.Stat(filepath.Join(dir, ".ssh", "known_hosts"))
		if !valid && !os.IsNotExist(statErr) {
			t.Fatal("unverified host key installed")
		}
	}
}
