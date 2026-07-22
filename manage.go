package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type managedPaste struct {
	ID                string `json:"id"`
	Filename          string `json:"filename,omitempty"`
	Label             string `json:"label,omitempty"`
	URL               string `json:"url"`
	CreatedAt         string `json:"created_at"`
	Expires           string `json:"expires,omitempty"`
	DataPolicy        string `json:"data_policy"`
	Size              int64  `json:"size"`
	ContentType       string `json:"content_type"`
	PasswordProtected bool   `json:"password_protected"`
}

type manageAction struct {
	Action      string `json:"action"`
	Label       string `json:"label,omitempty"`
	DataPolicy  string `json:"data_policy,omitempty"`
	NewPassword string `json:"new_password,omitempty"`
	Password    string `json:"password,omitempty"`
}

func (a application) runManage(args []string) int {
	if len(args) == 1 && (args[0] == "--help" || args[0] == "-h") {
		fmt.Fprint(a.stdout, manageUsageText)
		return 0
	}

	command, target, action, deleting, err := parseManageArguments(args)
	if err != nil {
		fmt.Fprint(a.stderr, manageUsageText)
		return 2
	}
	cfg, err := loadConfig(a.configPath)
	if err != nil {
		fmt.Fprintln(a.stderr, err)
		return 2
	}
	reference, err := resolvePrivatePasteReference(cfg.ServerURL, target, "manage")
	if err != nil {
		fmt.Fprintf(a.stderr, "invalid manage target: %v\n", err)
		return 2
	}
	if reference.token == "" {
		reference.token, err = a.promptSecret("Manage token: ", "manage token must not be empty")
		if err != nil {
			fmt.Fprintf(a.stderr, "cannot read manage token: %v\n", err)
			return 2
		}
	}

	if command == "password-enable" {
		action.NewPassword, err = a.promptNewPassword()
		if err != nil {
			fmt.Fprintf(a.stderr, "cannot read new password: %v\n", err)
			return 2
		}
	}
	if command == "password-disable" {
		action.Password, err = a.promptExistingPassword("Current paste password: ")
		if err != nil {
			fmt.Fprintf(a.stderr, "cannot read current password: %v\n", err)
			return 2
		}
	}

	if deleting {
		if err := deleteManagedPaste(a.requestContext(), a.httpClient, cfg, reference.code, reference.token); err != nil {
			fmt.Fprintln(a.stderr, err)
			return 1
		}
		if _, err := fmt.Fprintf(a.stdout, "deleted: %s\n", reference.code); err != nil {
			fmt.Fprintf(a.stderr, "write output: %v\n", err)
			return 1
		}
		return 0
	}

	method := http.MethodPatch
	var payload *manageAction
	if command == "show" {
		method = http.MethodGet
	} else {
		payload = &action
	}
	result, err := requestManagedPaste(a.requestContext(), a.httpClient, cfg, method, reference.code, reference.token, payload)
	if err != nil {
		fmt.Fprintln(a.stderr, err)
		return 1
	}
	if err := writeManagedPaste(a.stdout, result); err != nil {
		fmt.Fprintf(a.stderr, "write output: %v\n", err)
		return 1
	}
	return 0
}

func parseManageArguments(args []string) (command string, target string, action manageAction, deleting bool, err error) {
	switch {
	case len(args) == 2 && args[0] == "show":
		return "show", args[1], manageAction{}, false, nil
	case len(args) == 3 && args[0] == "label":
		return "label", args[1], manageAction{Action: "set_label", Label: args[2]}, false, nil
	case len(args) == 3 && args[0] == "policy":
		return "policy", args[1], manageAction{Action: "set_policy", DataPolicy: args[2]}, false, nil
	case len(args) == 3 && args[0] == "password" && args[1] == "enable":
		return "password-enable", args[2], manageAction{Action: "enable_password"}, false, nil
	case len(args) == 3 && args[0] == "password" && args[1] == "disable":
		return "password-disable", args[2], manageAction{Action: "disable_password"}, false, nil
	case len(args) == 2 && args[0] == "delete":
		return "delete", args[1], manageAction{}, true, nil
	default:
		return "", "", manageAction{}, false, fmt.Errorf("invalid manage arguments")
	}
}

func requestManagedPaste(ctx context.Context, client *http.Client, cfg config, method, code, token string, action *manageAction) (managedPaste, error) {
	var body io.Reader
	if action != nil {
		encoded, err := json.Marshal(action)
		if err != nil {
			return managedPaste{}, fmt.Errorf("encode manage request: %w", err)
		}
		body = bytes.NewReader(encoded)
	}
	endpoint := strings.TrimRight(cfg.ServerURL, "/") + "/api/v1/pastes/" + url.PathEscape(code)
	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return managedPaste{}, fmt.Errorf("create manage request: %w", err)
	}
	req.Header.Set("paste-manage-token", token)
	if action != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := doManagedRequest(client, cfg.ServerURL, token, req)
	if err != nil {
		return managedPaste{}, err
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
	if err != nil {
		return managedPaste{}, fmt.Errorf("read manage response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return managedPaste{}, responseError("manage failed", resp, responseBody)
	}
	var result managedPaste
	if err := json.Unmarshal(responseBody, &result); err != nil || strings.TrimSpace(result.ID) == "" {
		return managedPaste{}, fmt.Errorf("invalid manage response from server")
	}
	return result, nil
}

func deleteManagedPaste(ctx context.Context, client *http.Client, cfg config, code, token string) error {
	endpoint := strings.TrimRight(cfg.ServerURL, "/") + "/api/v1/pastes/" + url.PathEscape(code)
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, endpoint, nil)
	if err != nil {
		return fmt.Errorf("create manage delete request: %w", err)
	}
	req.Header.Set("paste-manage-token", token)
	resp, err := doManagedRequest(client, cfg.ServerURL, token, req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	responseBody, err := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
	if err != nil {
		return fmt.Errorf("read manage delete response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return responseError("manage delete failed", resp, responseBody)
	}
	var result struct {
		Deleted bool `json:"deleted"`
	}
	if err := json.Unmarshal(responseBody, &result); err != nil || !result.Deleted {
		return fmt.Errorf("invalid manage delete response from server")
	}
	return nil
}

func doManagedRequest(client *http.Client, serverURL, token string, req *http.Request) (*http.Response, error) {
	sensitiveHeaders := make(http.Header)
	sensitiveHeaders.Set("paste-manage-token", token)
	redirectClient, err := secureHTTPClient(client, serverURL, sensitiveHeaders)
	if err != nil {
		return nil, fmt.Errorf("configure manage redirects: %w", err)
	}
	resp, err := redirectClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("manage request failed for %s: %w", serverURL, err)
	}
	return resp, nil
}

func writeManagedPaste(output io.Writer, paste managedPaste) error {
	filename := paste.Filename
	if filename == "" {
		filename = "-"
	}
	label := paste.Label
	if label == "" {
		label = "-"
	}
	expires := paste.Expires
	if expires == "" {
		expires = "-"
	}
	protected := "no"
	if paste.PasswordProtected {
		protected = "yes"
	}
	_, err := fmt.Fprintf(output, "code: %s\nurl: %s\nfilename: %s\nlabel: %s\ncreated: %s\nexpires: %s\npolicy: %s\nsize: %d\ncontent-type: %s\npassword-protected: %s\n",
		paste.ID, paste.URL, filename, label, paste.CreatedAt, expires, paste.DataPolicy, paste.Size, paste.ContentType, protected)
	return err
}
