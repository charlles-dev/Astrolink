package gateway

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
)

type RouterDiagnostic struct {
	Online      bool
	OpenNDS     OpenNDSStatus
	Board       RouterBoard
	ClientCount int
	Clients     []ClientSummary
	RecentLogs  []string
}

type OpenNDSStatus struct {
	Online                   bool
	Version                  string
	Uptime                   string
	GatewayName              string
	ClientCount              int
	AuthenticatedClientCount int
}

type ClientSummary struct {
	MAC             string
	IP              string
	Token           string
	State           string
	DownloadedBytes int64
	UploadedBytes   int64
}

type RouterBoard struct {
	Hostname  string
	Model     string
	BoardName string
	System    string
	Kernel    string
	Firmware  string
}

var (
	ipPattern      = regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}\b`)
	inlineMACRegex = regexp.MustCompile(`\b[0-9A-F]{2}(?::[0-9A-F]{2}){5}\b`)
	numberPattern  = regexp.MustCompile(`\d+`)
	spacePattern   = regexp.MustCompile(`\s+`)
	offlineMarkers = []string{"not running", "offline", "connection refused", "unable to connect", "command not found", "not found"}
)

func ParseOpenNDSStatus(output string) OpenNDSStatus {
	status := OpenNDSStatus{}
	trimmed := strings.TrimSpace(output)
	if trimmed == "" {
		return status
	}

	lowerOutput := strings.ToLower(trimmed)
	status.Online = true
	for _, marker := range offlineMarkers {
		if strings.Contains(lowerOutput, marker) {
			status.Online = false
			break
		}
	}

	for _, rawLine := range strings.Split(output, "\n") {
		key, value, ok := splitKeyValue(rawLine)
		if !ok {
			continue
		}
		switch normalizeKey(key) {
		case "version", "openndsversion":
			status.Version = value
		case "uptime":
			status.Uptime = value
		case "gatewayname":
			status.GatewayName = value
		case "currentclients", "clients", "clientcount":
			status.ClientCount = parseFirstInt(value)
		case "authenticatedclients", "authenticatedclientcount":
			status.AuthenticatedClientCount = parseFirstInt(value)
		}
	}

	return status
}

func ParseOpenNDSClients(output string) []ClientSummary {
	var clients []ClientSummary
	var current ClientSummary

	flush := func() {
		if hasClientData(current) {
			clients = append(clients, current)
			current = ClientSummary{}
		}
	}

	for _, rawLine := range strings.Split(output, "\n") {
		line := strings.TrimSpace(rawLine)
		if line == "" {
			flush()
			continue
		}
		if strings.HasPrefix(strings.ToLower(line), "client ") {
			flush()
			continue
		}

		if key, value, ok := splitKeyValue(line); ok && isClientFieldKey(key) {
			applyClientField(&current, key, value)
			continue
		}

		if compact, ok := parseCompactClientLine(line); ok {
			flush()
			clients = append(clients, compact)
		}
	}
	flush()

	return clients
}

func ParseSystemBoard(output string) RouterBoard {
	var payload struct {
		Kernel    string `json:"kernel"`
		Hostname  string `json:"hostname"`
		System    string `json:"system"`
		Model     string `json:"model"`
		BoardName string `json:"board_name"`
		Release   struct {
			Distribution string `json:"distribution"`
			Version      string `json:"version"`
			Revision     string `json:"revision"`
		} `json:"release"`
	}
	if err := json.Unmarshal([]byte(output), &payload); err != nil {
		return RouterBoard{}
	}

	board := RouterBoard{
		Hostname:  payload.Hostname,
		Model:     payload.Model,
		BoardName: payload.BoardName,
		System:    payload.System,
		Kernel:    payload.Kernel,
	}
	board.Firmware = joinNonEmpty(" ", payload.Release.Distribution, payload.Release.Version, payload.Release.Revision)
	return board
}

func ParseOpenNDSLogs(output string, maxLines int) []string {
	var lines []string
	for _, rawLine := range strings.Split(output, "\n") {
		line := strings.TrimSpace(rawLine)
		if line != "" {
			lines = append(lines, line)
		}
	}
	if maxLines > 0 && len(lines) > maxLines {
		return lines[len(lines)-maxLines:]
	}
	return lines
}

func BuildRouterDiagnostic(statusOutput, clientsOutput, boardOutput, logsOutput string) RouterDiagnostic {
	status := ParseOpenNDSStatus(statusOutput)
	clients := ParseOpenNDSClients(clientsOutput)
	if status.ClientCount == 0 && len(clients) > 0 {
		status.ClientCount = len(clients)
	}

	return RouterDiagnostic{
		Online:      status.Online,
		OpenNDS:     status,
		Board:       ParseSystemBoard(boardOutput),
		ClientCount: status.ClientCount,
		Clients:     clients,
		RecentLogs:  ParseOpenNDSLogs(logsOutput, 50),
	}
}

func applyClientField(client *ClientSummary, key, value string) {
	switch normalizeKey(key) {
	case "mac", "macaddress", "clientmac":
		client.MAC = normalizeDiagnosticMAC(value)
	case "ip", "ipaddress", "clientip":
		client.IP = value
	case "token":
		client.Token = value
	case "state", "status":
		client.State = value
	case "download", "downloaded", "downloadedbytes":
		client.DownloadedBytes = parseFirstInt64(value)
	case "upload", "uploaded", "uploadedbytes":
		client.UploadedBytes = parseFirstInt64(value)
	}
}

func isClientFieldKey(key string) bool {
	switch normalizeKey(key) {
	case "mac", "macaddress", "clientmac", "ip", "ipaddress", "clientip", "token", "state", "status", "download", "downloaded", "downloadedbytes", "upload", "uploaded", "uploadedbytes":
		return true
	default:
		return false
	}
}

func parseCompactClientLine(line string) (ClientSummary, bool) {
	mac := inlineMACRegex.FindString(strings.ToUpper(strings.ReplaceAll(line, "-", ":")))
	ip := ipPattern.FindString(line)
	if mac == "" && ip == "" {
		return ClientSummary{}, false
	}

	client := ClientSummary{MAC: mac, IP: ip}
	for _, field := range strings.Fields(line) {
		cleaned := strings.Trim(field, ",;")
		if strings.EqualFold(cleaned, mac) || cleaned == ip || strings.Contains(cleaned, ":") {
			continue
		}
		if looksLikeClientState(cleaned) {
			client.State = cleaned
			break
		}
	}
	return client, true
}

func splitKeyValue(line string) (string, string, bool) {
	line = strings.TrimSpace(line)
	if line == "" {
		return "", "", false
	}
	for _, separator := range []string{":", "="} {
		if before, after, found := strings.Cut(line, separator); found {
			key := strings.TrimSpace(before)
			value := strings.TrimSpace(after)
			if key != "" && value != "" {
				return key, value, true
			}
		}
	}
	return "", "", false
}

func normalizeKey(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = strings.ReplaceAll(value, "_", "")
	value = strings.ReplaceAll(value, "-", "")
	return strings.ReplaceAll(value, " ", "")
}

func normalizeDiagnosticMAC(value string) string {
	value = strings.TrimSpace(value)
	value = strings.Trim(value, ",;")
	return strings.ToUpper(strings.ReplaceAll(value, "-", ":"))
}

func hasClientData(client ClientSummary) bool {
	return client.MAC != "" || client.IP != "" || client.Token != "" || client.State != ""
}

func looksLikeClientState(value string) bool {
	value = strings.ToLower(value)
	return strings.Contains(value, "auth") || strings.Contains(value, "client") || value == "trusted" || value == "blocked"
}

func parseFirstInt(value string) int {
	parsed, _ := strconv.Atoi(numberPattern.FindString(value))
	return parsed
}

func parseFirstInt64(value string) int64 {
	parsed, _ := strconv.ParseInt(numberPattern.FindString(value), 10, 64)
	return parsed
}

func joinNonEmpty(separator string, values ...string) string {
	var filtered []string
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			filtered = append(filtered, value)
		}
	}
	return spacePattern.ReplaceAllString(strings.Join(filtered, separator), " ")
}
