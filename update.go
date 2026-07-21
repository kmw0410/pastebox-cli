package main

import (
	"bufio"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

const defaultReleaseAPIURL = "https://api.github.com/repos/kmw0410/pastebox-cli/releases/latest"

type releaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
	Digest             string `json:"digest"`
}

type latestRelease struct {
	TagName string         `json:"tag_name"`
	Assets  []releaseAsset `json:"assets"`
}

type updateTarget struct {
	assetSuffix string
	installer   string
}

func (a application) runUpdate(args []string) int {
	if len(args) == 1 && (args[0] == "--help" || args[0] == "-h") {
		fmt.Fprint(a.stdout, updateUsageText)
		return 0
	}
	if len(args) != 0 {
		fmt.Fprint(a.stderr, updateUsageText)
		return 2
	}

	distro, err := readOSRelease(a.updateOSReleasePath())
	if err != nil {
		fmt.Fprintf(a.stderr, "cannot detect Linux distribution: %v\n", err)
		return 2
	}
	if distroMatches(distro, "arch") {
		return a.runArchUpdate()
	}

	target, err := selectUpdateTarget(distro, a.updateGOARCH())
	if err != nil {
		fmt.Fprintln(a.stderr, err)
		return 2
	}
	release, err := a.fetchLatestRelease()
	if err != nil {
		fmt.Fprintf(a.stderr, "cannot check for updates: %v\n", err)
		return 1
	}
	if release.TagName == version {
		fmt.Fprintf(a.stdout, "pastebox-cli is already up to date (%s).\n", version)
		return 0
	}

	asset, err := selectReleaseAsset(release.Assets, target.assetSuffix)
	if err != nil {
		fmt.Fprintf(a.stderr, "cannot select update package for %s: %v\n", release.TagName, err)
		return 1
	}
	packagePath, err := a.downloadVerifiedAsset(asset)
	if err != nil {
		fmt.Fprintf(a.stderr, "cannot download update: %v\n", err)
		return 1
	}
	defer os.Remove(packagePath)

	fmt.Fprintf(a.stdout, "Installing pastebox-cli %s from %s...\n", release.TagName, asset.Name)
	if err := a.installPackage(target, packagePath); err != nil {
		fmt.Fprintf(a.stderr, "cannot install update: %v\n", err)
		return 1
	}
	fmt.Fprintf(a.stdout, "Updated pastebox-cli to %s.\n", release.TagName)
	return 0
}

func (a application) runArchUpdate() int {
	release, err := a.fetchLatestRelease()
	if err != nil {
		fmt.Fprintf(a.stderr, "cannot check for updates: %v\n", err)
		return 1
	}
	if release.TagName == version {
		fmt.Fprintf(a.stdout, "pastebox-cli is already up to date (%s).\n", version)
		return 0
	}

	helper := a.findAURHelper()
	if helper == "" {
		fmt.Fprintln(a.stdout, "A newer pastebox-cli release is available, but neither paru nor yay is installed.")
		fmt.Fprintln(a.stdout, "Install paru or yay, then run pb update again.")
		return 0
	}

	fmt.Fprintf(a.stdout, "Updating pastebox-cli to %s with %s...\n", release.TagName, helper)
	if err := a.executeCommand(helper, "-S", "pastebox-cli"); err != nil {
		fmt.Fprintf(a.stderr, "cannot update with %s: %v\n", helper, err)
		return 1
	}
	return 0
}

func (a application) findAURHelper() string {
	lookPath := exec.LookPath
	if a.lookPath != nil {
		lookPath = a.lookPath
	}
	for _, name := range []string{"paru", "yay"} {
		if _, err := lookPath(name); err == nil {
			return name
		}
	}
	return ""
}

func (a application) executeCommand(name string, args ...string) error {
	if a.runCommand != nil {
		return a.runCommand(a.requestContext(), name, args...)
	}
	cmd := exec.CommandContext(a.requestContext(), name, args...)
	cmd.Stdin = a.stdin
	cmd.Stdout = a.stdout
	cmd.Stderr = a.stderr
	return cmd.Run()
}

func (a application) updateOSReleasePath() string {
	if a.osReleasePath != "" {
		return a.osReleasePath
	}
	return "/etc/os-release"
}

func (a application) updateGOARCH() string {
	if a.goarch != "" {
		return a.goarch
	}
	return runtime.GOARCH
}

func readOSRelease(path string) (map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	values := make(map[string]string)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		key, value, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		values[key] = strings.Trim(strings.TrimSpace(value), "\"'")
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return values, nil
}

func distroMatches(values map[string]string, names ...string) bool {
	known := strings.Fields(strings.ToLower(values["ID"] + " " + values["ID_LIKE"]))
	for _, value := range known {
		for _, name := range names {
			if value == name {
				return true
			}
		}
	}
	return false
}

func selectUpdateTarget(distro map[string]string, goarch string) (updateTarget, error) {
	if distroMatches(distro, "debian", "ubuntu") {
		switch goarch {
		case "amd64":
			return updateTarget{assetSuffix: "_amd64.deb", installer: "apt-get"}, nil
		case "arm64":
			return updateTarget{assetSuffix: "_arm64.deb", installer: "apt-get"}, nil
		default:
			return updateTarget{}, fmt.Errorf("Debian updates do not support architecture %s", goarch)
		}
	}
	if distroMatches(distro, "rhel", "fedora", "centos", "rocky", "almalinux") {
		if goarch != "amd64" {
			return updateTarget{}, fmt.Errorf("RPM updates do not support architecture %s", goarch)
		}
		return updateTarget{assetSuffix: ".x86_64.rpm", installer: "dnf"}, nil
	}
	return updateTarget{}, errors.New("automatic updates are supported only on Debian/Ubuntu and x86-64 RHEL/Fedora systems")
}

func (a application) fetchLatestRelease() (latestRelease, error) {
	apiURL := a.releaseAPIURL
	if apiURL == "" {
		apiURL = defaultReleaseAPIURL
	}
	req, err := http.NewRequestWithContext(a.requestContext(), http.MethodGet, apiURL, nil)
	if err != nil {
		return latestRelease{}, err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", "pastebox-cli/"+version)
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return latestRelease{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return latestRelease{}, fmt.Errorf("GitHub returned HTTP %d", resp.StatusCode)
	}
	var release latestRelease
	decoder := json.NewDecoder(io.LimitReader(resp.Body, 2<<20))
	if err := decoder.Decode(&release); err != nil {
		return latestRelease{}, fmt.Errorf("invalid GitHub response: %w", err)
	}
	if strings.TrimSpace(release.TagName) == "" {
		return latestRelease{}, errors.New("GitHub response is missing tag_name")
	}
	return release, nil
}

func selectReleaseAsset(assets []releaseAsset, suffix string) (releaseAsset, error) {
	var match *releaseAsset
	for i := range assets {
		if strings.HasPrefix(assets[i].Name, "pastebox-cli") && strings.HasSuffix(assets[i].Name, suffix) {
			if match != nil {
				return releaseAsset{}, fmt.Errorf("multiple assets match *%s", suffix)
			}
			match = &assets[i]
		}
	}
	if match == nil {
		return releaseAsset{}, fmt.Errorf("release does not contain an asset matching *%s", suffix)
	}
	return *match, nil
}

func (a application) downloadVerifiedAsset(asset releaseAsset) (string, error) {
	expected, err := parseSHA256Digest(asset.Digest)
	if err != nil {
		return "", fmt.Errorf("asset %s has invalid digest: %w", asset.Name, err)
	}
	if err := validateDownloadURL(asset.BrowserDownloadURL, a.releaseAPIURL != ""); err != nil {
		return "", err
	}
	req, err := http.NewRequestWithContext(a.requestContext(), http.MethodGet, asset.BrowserDownloadURL, nil)
	if err != nil {
		return "", err
	}
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("asset download returned HTTP %d", resp.StatusCode)
	}

	pattern := "pastebox-cli-update-*" + filepath.Ext(asset.Name)
	file, err := os.CreateTemp("", pattern)
	if err != nil {
		return "", err
	}
	path := file.Name()
	keep := false
	defer func() {
		file.Close()
		if !keep {
			os.Remove(path)
		}
	}()

	hash := sha256.New()
	if _, err := io.Copy(io.MultiWriter(file, hash), resp.Body); err != nil {
		return "", err
	}
	if err := file.Close(); err != nil {
		return "", err
	}
	actual := hash.Sum(nil)
	if subtle.ConstantTimeCompare(actual, expected) != 1 {
		return "", fmt.Errorf("SHA-256 mismatch for %s", asset.Name)
	}
	keep = true
	return path, nil
}

func parseSHA256Digest(value string) ([]byte, error) {
	algorithm, encoded, ok := strings.Cut(strings.TrimSpace(value), ":")
	if !ok || !strings.EqualFold(algorithm, "sha256") {
		return nil, errors.New("expected sha256 digest")
	}
	digest, err := hex.DecodeString(encoded)
	if err != nil || len(digest) != sha256.Size {
		return nil, errors.New("expected 64 hexadecimal characters")
	}
	return digest, nil
}

func validateDownloadURL(rawURL string, allowTestServer bool) error {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid asset download URL: %w", err)
	}
	if allowTestServer {
		if parsed.Scheme != "http" && parsed.Scheme != "https" {
			return errors.New("invalid asset download URL scheme")
		}
		return nil
	}
	if parsed.Scheme != "https" || !strings.EqualFold(parsed.Hostname(), "github.com") || parsed.User != nil {
		return errors.New("asset download URL is not an HTTPS GitHub URL")
	}
	if !strings.HasPrefix(parsed.EscapedPath(), "/kmw0410/pastebox-cli/releases/download/") {
		return errors.New("asset download URL does not belong to the pastebox-cli repository")
	}
	return nil
}

func (a application) installPackage(target updateTarget, packagePath string) error {
	args := []string{"install", "-y", packagePath}
	name := target.installer
	effectiveUID := os.Geteuid
	if a.effectiveUID != nil {
		effectiveUID = a.effectiveUID
	}
	if effectiveUID() != 0 {
		args = append([]string{name}, args...)
		name = "sudo"
	}
	return a.executeCommand(name, args...)
}
