package notification

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

const alertzyURL = "https://alertzy.app/send"

var ApiKeys []string

// Priority levels for notifications
const (
	PriorityNormal   = 0
	PriorityHigh     = 1
	PriorityCritical = 2
)

// Button represents an actionable button in a notification
type Button struct {
	Text  string `json:"text"`
	Link  string `json:"link"`
	Color string `json:"color"`
}

// AlertzyRequest represents the full request to Alertzy API
type AlertzyRequest struct {
	AccountKey string   `json:"accountKey"`
	Title      string   `json:"title,omitempty"`
	Message    string   `json:"message,omitempty"`
	Priority   int      `json:"priority,omitempty"`
	Group      string   `json:"group,omitempty"`
	Image      string   `json:"image,omitempty"`
	Link       string   `json:"link,omitempty"`
	Buttons    []Button `json:"buttons,omitempty"`
}

// AlertzyResponse represents the API response
type AlertzyResponse struct {
	Response string            `json:"response"`
	SentTo   []string          `json:"sentTo"`
	Error    map[string]string `json:"error,omitempty"`
}

// SetKeys sets the default account keys for notifications, filtering out empty strings.
func SetKeys(keys []string) {

	var filtered []string
	for _, s := range keys {
		if s != "" {
			filtered = append(filtered, s)
		}
	}

	ApiKeys = filtered

}

// Send sends a notification to all default keys.
func Send(title, message string) error {
	if len(ApiKeys) == 0 {
		return fmt.Errorf("no default keys set, use SetKeys() first")
	}

	accountKey := strings.Join(ApiKeys, "_")
	return SendTo(accountKey, title, message)
}

// SendTo sends a notification to specific account key(s).
func SendTo(accountKey, title, message string) error {
	req := AlertzyRequest{
		AccountKey: accountKey,
		Title:      title,
		Message:    message,
	}
	return SendRequest(req)
}

// SendWithOptions sends a notification with full options
func SendWithOptions(accountKey, title, message string, priority int, group, imageURL, link string, buttons []Button) error {
	req := AlertzyRequest{
		AccountKey: accountKey,
		Title:      title,
		Message:    message,
		Priority:   priority,
		Group:      group,
		Image:      imageURL,
		Link:       link,
		Buttons:    buttons,
	}
	return SendRequest(req)
}

// SendRequest sends a full AlertzyRequest to the API and handles the response.
//
// It prepares the request data based on the fields of the AlertzyRequest struct, including account key, title, message, priority, group, image, link, and buttons. It then creates an HTTP POST request to the alertzyURL, sends the request, and reads the response. The response is parsed into an AlertzyResponse struct, and any errors are returned if the response indicates a failure or mixed results.
func SendRequest(req AlertzyRequest) error {
	// Prepare form data
	data := url.Values{}
	data.Set("accountKey", req.AccountKey)

	if req.Title != "" {
		data.Set("title", req.Title)
	}
	if req.Message != "" {
		data.Set("message", req.Message)
	}
	if req.Priority > 0 {
		data.Set("priority", fmt.Sprintf("%d", req.Priority))
	}
	if req.Group != "" {
		data.Set("group", req.Group)
	}
	if req.Image != "" {
		data.Set("image", req.Image)
	}
	if req.Link != "" {
		data.Set("link", req.Link)
	}
	if len(req.Buttons) > 0 {
		buttonsJSON, err := json.Marshal(req.Buttons)
		if err != nil {
			return fmt.Errorf("failed to marshal buttons: %w", err)
		}
		data.Set("buttons", string(buttonsJSON))
	}

	// Create HTTP request
	httpReq, err := http.NewRequest("POST", alertzyURL, bytes.NewBufferString(data.Encode()))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Send request
	client := &http.Client{}
	resp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Parse response
	var alertzyResp AlertzyResponse
	if err := json.Unmarshal(body, &alertzyResp); err != nil {
		return fmt.Errorf("failed to parse response: %w", err)
	}

	// Check response status
	if alertzyResp.Response == "fail" {
		var errors []string
		for key, msg := range alertzyResp.Error {
			errors = append(errors, fmt.Sprintf("%s: %s", key, msg))
		}
		return fmt.Errorf("alertzy error: %s", strings.Join(errors, ", "))
	}

	if alertzyResp.Response == "mixed" {
		var errors []string
		for key, msg := range alertzyResp.Error {
			errors = append(errors, fmt.Sprintf("%s: %s", key, msg))
		}
		fmt.Printf("Warning: some notifications failed: %s\n", strings.Join(errors, ", "))
	}

	return nil
}
