package main

import (
	"os"
	"strings"
	"testing"
)

func TestReleaseBuildRunsQualityChecksForPushAndPullRequest(t *testing.T) {
	workflow := readWorkflow(t, ".github/workflows/release-build.yml")
	for _, want := range []string{
		"push:\n    paths-ignore:\n      - \"**/*.md\"\n      - \".gitignore\"\n      - \".github/workflows/**\"\n      - \"**/LICENSE\"",
		"pull_request:\n    paths-ignore:\n      - \"**/*.md\"\n      - \".gitignore\"\n      - \".github/workflows/**\"\n      - \"**/LICENSE\"",
		"go-version-file: go.mod",
		"cache: true",
		"unformatted=\"$(gofmt -l .)\"",
		"go vet ./...",
		"go test ./...",
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
}

func readWorkflow(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
