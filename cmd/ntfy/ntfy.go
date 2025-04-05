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
	NtfyServer          	string 				`json:"ntfy_server"`
	Username         		string 				`json:"username"`
	Password		 		string 				`json:"password"`
	AccessToken				string 				`json:"access_token"`
	ResponseStatusCodes 	string 				`json:"response_status_codes"`
	IcingaServerUrl 		string 				`json:"icinga_server_url"`
	AlertOnAckClear			bool   				`json:"alert_on_ack_clear"`
	AlertOnAckSet			bool   				`json:"alert_on_ack_set"`
	AlertOnCustom			bool   				`json:"alert_on_custom"`
	AlertOnDowntimeEnd		bool   				`json:"alert_on_downtime_end"`
	AlertOnDowntimeRemoved	bool   				`json:"alert_on_downtime_removed"`
	AlertOnDowntimeStart	bool   				`json:"alert_on_downtime_start"`
	AlertOnFlappingEnd		bool   				`json:"alert_on_flapping_end"`
	AlertOnFlappingStart	bool   				`json:"alert_on_flapping_start"`
	AlertOnIncidentAge		bool   				`json:"alert_on_incident_age"`
	AlertOnMute				bool   				`json:"alert_on_mute"`
	AlertOnIncidentState	bool   				`json:"alert_on_state"`
	AlertOnUnmute			bool   				`json:"alert_on_unmute"`
	DefaultPriority			string				`json:"default_priority"`
	PriorityMaxEvents 		map[string]string 	`json:"priority_max_events"`
	PriorityUrgentEvents 	map[string]string 	`json:"priority_max_events"`
	PriorityDefaultEvents 	map[string]string 	`json:"priority_max_events"`
	PriorityLowEvents 		map[string]string 	`json:"priority_max_events"`
	PriorityMinEvents 		map[string]string 	`json:"priority_max_events"`
	

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
			Type: "secret",
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
			Type: "secret",
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
		{
			Name: "icinga_server_url",
			Type: "string",
			Label: map[string]string{
				"en_US": "Icinga Server URL",
			},
			Help: map[string]string{
				"en_US": "Server URL for links, e.g., https://icinga.example.com",
			},
			Required: false,
		},
		{
			Name: "alert_on_ack_clear",
			Type: "bool",
			Label: map[string]string{
				"en_US": "Alert on Ack Clear",
			},
			Help: map[string]string{
				"en_US": "Sends an alert whenever an acknowledgement is removed",
			},
			Default: false,
		},
				{
			Name: "alert_on_ack_set",
			Type: "bool",
			Label: map[string]string{
				"en_US": "Alert on Ack Set",
			},
			Help: map[string]string{
				"en_US": "Sends an alert whenever an acknowledgement is added",
			},
			Default: false,
		},
				{
			Name: "alert_on_custom",
			Type: "bool",
			Label: map[string]string{
				"en_US": "Alert on Custom Event",
			},
			Help: map[string]string{
				"en_US": "Sends an alert whenever a custom event is thrown",
			},
			Default: false,
		},
				{
			Name: "alert_on_downtime_end",
			Type: "bool",
			Label: map[string]string{
				"en_US": "Alert on Downtime End",
			},
			Help: map[string]string{
				"en_US": "Sends an alert whenever downtime ends",
			},
			Default: false,
		},
				{
			Name: "alert_on_downtime_removed",
			Type: "bool",
			Label: map[string]string{
				"en_US": "Alert on Downtime Removed",
			},
			Help: map[string]string{
				"en_US": "Sends an alert whenever downtime is removed",
			},
			Default: false,
		},
				{
			Name: "alert_on_downtime_start",
			Type: "bool",
			Label: map[string]string{
				"en_US": "Alert on Downtime Start",
			},
			Help: map[string]string{
				"en_US": "Sends an alert whenever downtime starts",
			},
			Default: false,
		},
				{
			Name: "alert_on_flapping_end",
			Type: "bool",
			Label: map[string]string{
				"en_US": "Alert on Flapping End",
			},
			Help: map[string]string{
				"en_US": "Sends an alert whenever flapping stops",
			},
			Default: false,
		},
				{
			Name: "alert_on_flapping_start",
			Type: "bool",
			Label: map[string]string{
				"en_US": "Alert on Flapping Start",
			},
			Help: map[string]string{
				"en_US": "Sends an alert whenever flapping starts",
			},
			Default: false,
		},
				{
			Name: "alert_on_incident_age",
			Type: "bool",
			Label: map[string]string{
				"en_US": "Alert on Incident Age",
			},
			Help: map[string]string{
				"en_US": "Sends an alert whenever an incident age event is thrown",
			},
			Default: false,
		},
				{
			Name: "alert_on_mute",
			Type: "bool",
			Label: map[string]string{
				"en_US": "Alert on Incident Mute",
			},
			Help: map[string]string{
				"en_US": "Sends an alert whenever an incident is muted",
			},
			Default: false,
		},
				{
			Name: "alert_on_state",
			Type: "bool",
			Label: map[string]string{
				"en_US": "Alert on State",
			},
			Help: map[string]string{
				"en_US": "Sends an alert whenever a state event is thrown (default alert)",
			},
			Default: true,
		},
				{
			Name: "alert_on_unmute",
			Type: "bool",
			Label: map[string]string{
				"en_US": "Alert on Incident Unmute",
			},
			Help: map[string]string{
				"en_US": "Sends an alert whenever an incident is unmuted",
			},
			Default: false,
		},
		{
			Name: "default_priority",
			Type: "option",
			Label: map[string]string{
				"en_US": "Events to alert at max priority",
			},
			Help: map[string]string{
				"en_US": "When these events are thrown, set the notification to urgent priority",
			},
			Required: true,
			Options: map[string]string {
				"5": "max",
				"4": "high",
				"3": "default",
				"2": "low",
				"1": "min",
			},
			Default: "3",
		},
		{
			Name: "priority_max_events",
			Type: "options",
			Label: map[string]string{
				"en_US": "Events to alert at max priority",
			},
			Help: map[string]string{
				"en_US": "When these events are thrown, set the notification to urgent priority",
			},
			Options: map[string]string{
				"acknowledgement-cleared": "Acknowledgement Cleared",
				"acknowledgement-set": "Acknowledgement Set",
				"custom": "Custom",
				"downtime-end": "Downtime End",
				"downtime-start": "Downtime Start",
				"flapping-end": "Flapping End",
				"flapping-start": "Flapping Start",
				"incident-age": "Incident Age",
				"mute": "Mute",
				"state": "State",
				"unmute": "Unmute",
			},
		},
				{
			Name: "priority_high_events",
			Type: "options",
			Label: map[string]string{
				"en_US": "Events to alert at high priority",
			},
			Help: map[string]string{
				"en_US": "When these events are thrown, set the notification to high priority",
			},
			Options: map[string]string{
				"acknowledgement-cleared": "Acknowledgement Cleared",
				"acknowledgement-set": "Acknowledgement Set",
				"custom": "Custom",
				"downtime-end": "Downtime End",
				"downtime-start": "Downtime Start",
				"flapping-end": "Flapping End",
				"flapping-start": "Flapping Start",
				"incident-age": "Incident Age",
				"mute": "Mute",
				"state": "State",
				"unmute": "Unmute",
			},
		},
				{
			Name: "priority_default_events",
			Type: "options",
			Label: map[string]string{
				"en_US": "Events to alert at default priority",
			},
			Help: map[string]string{
				"en_US": "When these events are thrown, set the notification to default priority",
			},
			Options: map[string]string{
				"acknowledgement-cleared": "Acknowledgement Cleared",
				"acknowledgement-set": "Acknowledgement Set",
				"custom": "Custom",
				"downtime-end": "Downtime End",
				"downtime-start": "Downtime Start",
				"flapping-end": "Flapping End",
				"flapping-start": "Flapping Start",
				"incident-age": "Incident Age",
				"mute": "Mute",
				"state": "State",
				"unmute": "Unmute",
			},
		},
				{
			Name: "priority_low_events",
			Type: "options",
			Label: map[string]string{
				"en_US": "Events to alert at low priority",
			},
			Help: map[string]string{
				"en_US": "When these events are thrown, set the notification to low priority",
			},
			Options: map[string]string{
				"acknowledgement-cleared": "Acknowledgement Cleared",
				"acknowledgement-set": "Acknowledgement Set",
				"custom": "Custom",
				"downtime-end": "Downtime End",
				"downtime-start": "Downtime Start",
				"flapping-end": "Flapping End",
				"flapping-start": "Flapping Start",
				"incident-age": "Incident Age",
				"mute": "Mute",
				"state": "State",
				"unmute": "Unmute",
			},
		},
				{
			Name: "priority_min_events",
			Type: "options",
			Label: map[string]string{
				"en_US": "Events to alert at min priority",
			},
			Help: map[string]string{
				"en_US": "When these events are thrown, set the notification to min priority",
			},
			Options: map[string]string{
				"acknowledgement-cleared": "Acknowledgement Cleared",
				"acknowledgement-set": "Acknowledgement Set",
				"custom": "Custom",
				"downtime-end": "Downtime End",
				"downtime-start": "Downtime Start",
				"flapping-end": "Flapping End",
				"flapping-start": "Flapping Start",
				"incident-age": "Incident Age",
				"mute": "Mute",
				"state": "State",
				"unmute": "Unmute",
			},
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

	if string(jsonStr) != "[]" {
		err = json.Unmarshal(jsonStr, ch)
		if err != nil {
			return err
		}
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
		if strings.ToLower(addr.Type) == "ntfy" {
			ntfyAddress = addr.Address
			break
		}
	}

	// Verify address not blank
	if ntfyAddress == "" {
		return fmt.Errorf("Ntfy.sh address must not be blank.")
	}

	var severityFmt string
	var emojiTags string
	var priority string

	switch severity := req.Incident.Severity; severity {
	case "crit":
		if req.Object.Tags["service"] == "" {
			severityFmt = "DOWN"
		} else {
			severityFmt = "CRITICAL"
		}
		
		emojiTags = "bangbang"
		priority = "urgent"
	case "warning":
		severityFmt = "WARNING"
		emojiTags = "warning"
		priority = "high"
	case "ok":
		if req.Object.Tags["service"] == "" {
			severityFmt = "UP"
		} else {
			severityFmt = "OK"
		}
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

	var body string
	
	// switch eventType := req.Event.Type; eventType {
	// 	case "acknowledgement-cleared":

	// 	case "acknowledgement-set":

	// 	case "custom":

	// 	case "downtime-end":

	// 	case "downtime-removed":

	// 	case "downtime-start":

	// 	case "flapping-end":

	// 	case "flapping-start":

	// 	case "incident-age":

	// 	case "mute":

	// 	case "state":

	// 	case "unmute":

	// 	case default:

	// }
	body = "```\n" + req.Event.Message + "\n```"

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

	if severityFmt == "UP" || severityFmt == "DOWN" {
		httpReq.Header.Set("Title", req.Object.Tags["host"] + " is " + severityFmt)
	} else {
		httpReq.Header.Set("Title", req.Object.Tags["service"] + " on " + req.Object.Tags["host"] + " is " + severityFmt)
	}
	httpReq.Header.Set("Priority", priority)
	httpReq.Header.Set("Tags", emojiTags)
	httpReq.Header.Set("Action", "view,Open Icinga," + strings.Replace(req.Object.Url, "localhost", ch.IcingaServerUrl, -1))
	httpReq.Header.Set("Markdown", "true")
	

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