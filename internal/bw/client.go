package bw

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
)

type Client struct {
	session string
}

func NewClient(session string) *Client {
	return &Client{session: session}
}

func (c *Client) GetStatus() (*Status, error) {
	out, err := c.run("status")
	if err != nil {
		return nil, fmt.Errorf("getting status: %w", err)
	}
	var status Status
	if err := json.Unmarshal(out, &status); err != nil {
		return nil, fmt.Errorf("parsing status: %w", err)
	}
	return &status, nil
}

func (c *Client) Unlock(password string) (string, error) {
	cmd := exec.Command("bw", "unlock", "--raw", "--passwordenv", "BW_PASSWORD")
	cmd.Env = append(cmd.Environ(), "BW_PASSWORD="+password)
	if c.session != "" {
		cmd.Env = append(cmd.Env, "BW_SESSION="+c.session)
	}
	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("unlocking vault: %w", err)
	}
	session := strings.TrimSpace(string(out))
	c.session = session
	return session, nil
}

func (c *Client) Sync() error {
	_, err := c.run("sync")
	if err != nil {
		return fmt.Errorf("syncing vault: %w", err)
	}
	return nil
}

func (c *Client) ListItems() ([]Item, error) {
	out, err := c.run("list", "items")
	if err != nil {
		return nil, fmt.Errorf("listing items: %w", err)
	}
	var items []Item
	if err := json.Unmarshal(out, &items); err != nil {
		return nil, fmt.Errorf("parsing items: %w", err)
	}
	return items, nil
}

func (c *Client) ListFolders() ([]Folder, error) {
	out, err := c.run("list", "folders")
	if err != nil {
		return nil, fmt.Errorf("listing folders: %w", err)
	}
	var folders []Folder
	if err := json.Unmarshal(out, &folders); err != nil {
		return nil, fmt.Errorf("parsing folders: %w", err)
	}
	return folders, nil
}

func (c *Client) CreateFolder(name string) (*Folder, error) {
	payload := map[string]string{"name": name}
	encoded, err := encodeJSON(payload)
	if err != nil {
		return nil, fmt.Errorf("encoding folder: %w", err)
	}
	out, err := c.runWithPayload("create", "folder", encoded)
	if err != nil {
		return nil, fmt.Errorf("creating folder: %w", err)
	}
	var folder Folder
	if err := json.Unmarshal(out, &folder); err != nil {
		return nil, fmt.Errorf("parsing created folder: %w", err)
	}
	return &folder, nil
}

func (c *Client) EditFolder(id, name string) (*Folder, error) {
	payload := map[string]string{"id": id, "name": name}
	encoded, err := encodeJSON(payload)
	if err != nil {
		return nil, fmt.Errorf("encoding folder: %w", err)
	}
	out, err := c.runWithPayload("edit", "folder", id, encoded)
	if err != nil {
		return nil, fmt.Errorf("editing folder: %w", err)
	}
	var folder Folder
	if err := json.Unmarshal(out, &folder); err != nil {
		return nil, fmt.Errorf("parsing edited folder: %w", err)
	}
	return &folder, nil
}

func (c *Client) GetItem(id string) (*Item, error) {
	out, err := c.run("get", "item", id)
	if err != nil {
		return nil, fmt.Errorf("getting item: %w", err)
	}
	var item Item
	if err := json.Unmarshal(out, &item); err != nil {
		return nil, fmt.Errorf("parsing item: %w", err)
	}
	return &item, nil
}

func (c *Client) EditItem(item *Item) (*Item, error) {
	encoded, err := encodeJSON(item)
	if err != nil {
		return nil, fmt.Errorf("encoding item: %w", err)
	}
	out, err := c.runWithPayload("edit", "item", item.ID, encoded)
	if err != nil {
		return nil, fmt.Errorf("editing item: %w", err)
	}
	var updated Item
	if err := json.Unmarshal(out, &updated); err != nil {
		return nil, fmt.Errorf("parsing edited item: %w", err)
	}
	return &updated, nil
}

func (c *Client) DeleteItem(id string) error {
	_, err := c.run("delete", "item", id)
	if err != nil {
		return fmt.Errorf("deleting item: %w", err)
	}
	return nil
}

func (c *Client) DeleteFolder(id string) error {
	_, err := c.run("delete", "folder", id)
	if err != nil {
		return fmt.Errorf("deleting folder: %w", err)
	}
	return nil
}

func encodeJSON(v any) (string, error) {
	data, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(data), nil
}

func (c *Client) runWithPayload(args ...string) ([]byte, error) {
	// Last arg is the encoded payload, passed as final positional arg to bw
	displayArgs := make([]string, len(args))
	copy(displayArgs, args)
	displayArgs[len(displayArgs)-1] = "<payload>"

	return c.exec(args, displayArgs)
}

func (c *Client) run(args ...string) ([]byte, error) {
	return c.exec(args, args)
}

func (c *Client) exec(args, displayArgs []string) ([]byte, error) {
	cmd := exec.Command("bw", args...) //nolint:gosec // args are constructed internally, not from user input
	if c.session != "" {
		cmd.Env = append(cmd.Environ(), "BW_SESSION="+c.session)
	}
	out, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return nil, fmt.Errorf("bw %s: %s", strings.Join(displayArgs, " "), string(exitErr.Stderr))
		}
		return nil, err
	}
	return out, nil
}
