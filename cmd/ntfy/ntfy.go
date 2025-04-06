package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strconv"
	"strings"

	"github.com/icinga/icinga-notifications/pkg/plugin"
)

func main() {
	plugin.RunPlugin(&Ntfy{})
}

type Ntfy struct {
	NtfyServer          string `json:"ntfy_server"`
	Username         	string `json:"username"`
	Password		 	string `json:"password"`
	AccessToken			string `json:"access_token"`
	ResponseStatusCodes string `json:"response_status_codes"`

	respStatusCodes []int
}

func (ch *Ntfy) GetInfo() *plugin.Info {
	configAttrs := plugin.ConfigOptions{
		{
			Name: "ntfy_server",
			Type: "string",
			Label: map[string]string{
				"en_US": "Ntfy Server",
			},
			Help: map[string]string{
				"en_US": "Ntfy.sh server (default https://ntfy.sh)",
			},
			Default:  "https://ntfy.sh/",
			Required: true,
		},
		{
			Name: "username",
			Type: "string",
			Label: map[string]string{
				"en_US": "Username",
			},
			Help: map[string]string{
				"en_US": "Username to authenticate with (optional).",
			},
			Required: false,
		},
		{
			Name: "password",
			Type: "string",
			Label: map[string]string{
				"en_US": "Password",
			},
			Help: map[string]string{
				"en_US": "Password to authenticate with (optional).",
			},
			Required: false,
		},
		{
			Name: "access_token",
			Type: "string",
			Label: map[string]string{
				"en_US": "Access token",
			},
			Help: map[string]string{
				"en_US": "Access token to authenticate with (optional).",
			},
			Required: false,
		},
		{
			Name: "response_status_codes",
			Type: "string",
			Label: map[string]string{
				"en_US": "Response Status Codes",
			},
			Help: map[string]string{
				"en_US": "Comma separated list of expected HTTP response status code, e.g., 200,201,202,208,418",
			},
			Default:  "200",
			Required: true,
		},
	}

	return &plugin.Info{
		Name:             "Ntfy.sh",
		Version:          "1.0",
		Author:           "SourDusk",
		ConfigAttributes: configAttrs,
	}
}

func (ch *Ntfy) SetConfig(jsonStr json.RawMessage) error {
	err := plugin.PopulateDefaults(ch)
	if err != nil {
		return err
	}

	err = json.Unmarshal(jsonStr, ch)
	if err != nil {
		return err
	}

	// Verify that URL is a valid URL
	_, err = url.ParseRequestURI(ch.NtfyServer)
	if err != nil {
		return fmt.Errorf("%q is not a valid URI", ch.NtfyServer)
	}

	// Make sure server URL ends with a trailing /
	if !strings.HasSuffix(ch.NtfyServer, "/") {
		ch.NtfyServer = ch.NtfyServer + "/"
	}

	// Make sure both username and password are either filled or not
	if !((ch.Username == "" && ch.Password == "") || (ch.AccessToken != "" && ch.Password != "")) {
		return fmt.Errorf("Username and password must either both be empty or both be filled out")
	}

	// Make sure not using both username and password or access token
	if (ch.Username != "" || ch.Password != "") && ch.AccessToken != "" {
		return fmt.Errorf("Must have username and password or access token")
	}

	// Parse response status codes
	respStatusCodes := strings.Split(ch.ResponseStatusCodes, ",")
	ch.respStatusCodes = make([]int, len(respStatusCodes))
	for i, respStatusCodeStr := range respStatusCodes {
		respStatusCode, err := strconv.Atoi(respStatusCodeStr)
		if err != nil {
			return fmt.Errorf("cannot convert status code %q to int: %w", respStatusCodeStr, err)
		}
		ch.respStatusCodes[i] = respStatusCode
	}

	return nil
}

func (ch *Ntfy) SendNotification(req *plugin.NotificationRequest) error {
	// Get ntfy address from user's list
	var ntfyAddress string = ""
	for _, addr := range(req.Contact.Addresses) {
		if addr.Type == "ntfy" {
			ntfyAddress = addr.Address
			break
		}
	}

	// Verify address not blank
	if ntfyAddress == "" {
		return fmt.Errorf("Ntfy.sh address must not be blank.")
	}

	
	var emojiTags string
	var priority string
	var body string
	var title string

	// Various event types
	if req.Event.Type == "acknowledgement-set" || req.Event.Type == "downtime-start" {
		priority = "default"
		emojiTags = "information_source," + req.Event.Type
		var alertType string
		if req.Event.Type == "acknowledgement-set" {
			alertType = "Acknowledgement Set"
		} else {
			alertType = "Downtime Started"
		}
		title = alertType + " on " + req.Object.Name
		body = "Acknowledged by " + req.Contact.FullName + "\n" + req.Event.Message
	} else if req.Event.Type == "downtime-end" {
		priority = "high"
		emojiTags = "information_source," + req.Event.Type
		title = "Downtime Ended on " + req.Object.Name
	} else if req.Event.Type == "state" {
		var severityFmt string
		switch severity := req.Incident.Severity; severity {
		case "crit":
			severityFmt = "CRITICAL"
			emojiTags = "bangbang"
			priority = "urgent"
		case "warning":
			severityFmt = "WARNING"
			emojiTags = "warning"
			priority = "high"
		case "ok":
			severityFmt = "OK"
			emojiTags = "white_check_mark"
			priority = "default"
		case "down":
			severityFmt = "DOWN"
			emojiTags = "bangbang"
			priority = "urgent"
		default:
			severityFmt = "UNKNOWN"
			emojiTags = "question"
			priority = "urgent"
		}
		title = req.Object.Tags["service"] + " on " + req.Object.Tags["host"] + " is " + severityFmt
		body = "```" + req.Event.Message + "```"
	} else {
		return nil
	}
	

	httpReq, err := http.NewRequest("POST", ch.NtfyServer + ntfyAddress, strings.NewReader(body))
	if err != nil {
		return err
	}

	// Process authorization
	if ch.Username != "" { // Using username and password
		var authHeader string = "Basic " + base64.StdEncoding.EncodeToString([]byte(ch.Username + ":" + ch.Password))
		httpReq.Header.Set("Authorization", authHeader)
	} else if ch.AccessToken != "" {
		httpReq.Header.Set("Authorization", "Bearer " + ch.AccessToken)
	}

	httpReq.Header.Set("Title", title)
	httpReq.Header.Set("Priority", priority)
	httpReq.Header.Set("Tags", emojiTags)
	httpReq.Header.Set("Action", "view,Open Icinga," + strings.Replace(req.Object.Url, "localhost", "mon.gelat.in", -1))
	httpReq.Header.Set("Markdown", "yes")

	httpResp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return err
	}
	_, _ = io.Copy(io.Discard, httpResp.Body)
	_ = httpResp.Body.Close()

	if !slices.Contains(ch.respStatusCodes, httpResp.StatusCode) {
		return fmt.Errorf("unaccepted HTTP response status code %d not in %v",
			httpResp.StatusCode, ch.respStatusCodes)
	}

	return nil
}