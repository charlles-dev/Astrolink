package gateway_test

import (
	"testing"

	"github.com/astrolink/node/internal/gateway"
)

func TestParseOpenNDSStatus_ExtractsUsefulFields(t *testing.T) {
	status := gateway.ParseOpenNDSStatus(`
openNDS Status
Version: 10.2.0
Uptime: 3 days, 04:05:06
Gateway Name: Astrolink Cafe
Current clients: 2
Authenticated clients: 1
`)

	if !status.Online {
		t.Fatal("expected status to be online")
	}
	if status.Version != "10.2.0" {
		t.Fatalf("version = %q, want 10.2.0", status.Version)
	}
	if status.Uptime != "3 days, 04:05:06" {
		t.Fatalf("uptime = %q", status.Uptime)
	}
	if status.GatewayName != "Astrolink Cafe" {
		t.Fatalf("gateway name = %q", status.GatewayName)
	}
	if status.ClientCount != 2 {
		t.Fatalf("client count = %d, want 2", status.ClientCount)
	}
	if status.AuthenticatedClientCount != 1 {
		t.Fatalf("authenticated client count = %d, want 1", status.AuthenticatedClientCount)
	}
}

func TestParseOpenNDSStatus_ToleratesMissingFieldsAndOfflineText(t *testing.T) {
	status := gateway.ParseOpenNDSStatus("openNDS is not running\n")

	if status.Online {
		t.Fatal("expected status to be offline")
	}
	if status.Version != "" {
		t.Fatalf("version = %q, want empty", status.Version)
	}
}

func TestParseOpenNDSClients_ParsesBlocks(t *testing.T) {
	clients := gateway.ParseOpenNDSClients(`
Client 0
  MAC: AA:BB:CC:DD:EE:FF
  IP: 192.168.1.23
  Token: abc123
  State: Authenticated
  Downloaded: 12345
  Uploaded: 6789

Client 1
  MAC Address: 11:22:33:44:55:66
  IP Address: 192.168.1.24
  State: Preauthenticated
`)

	if len(clients) != 2 {
		t.Fatalf("len(clients) = %d, want 2", len(clients))
	}
	if clients[0].MAC != "AA:BB:CC:DD:EE:FF" || clients[0].IP != "192.168.1.23" {
		t.Fatalf("first client = %+v", clients[0])
	}
	if clients[0].State != "Authenticated" || clients[0].Token != "abc123" {
		t.Fatalf("first client = %+v", clients[0])
	}
	if clients[0].DownloadedBytes != 12345 || clients[0].UploadedBytes != 6789 {
		t.Fatalf("first client bytes = %+v", clients[0])
	}
	if clients[1].MAC != "11:22:33:44:55:66" || clients[1].State != "Preauthenticated" {
		t.Fatalf("second client = %+v", clients[1])
	}
}

func TestParseOpenNDSClients_ParsesCompactLines(t *testing.T) {
	clients := gateway.ParseOpenNDSClients(`
AA:BB:CC:DD:EE:FF 192.168.1.23 authenticated
11:22:33:44:55:66 192.168.1.24 preauthenticated
`)

	if len(clients) != 2 {
		t.Fatalf("len(clients) = %d, want 2", len(clients))
	}
	if clients[0].MAC != "AA:BB:CC:DD:EE:FF" || clients[0].IP != "192.168.1.23" {
		t.Fatalf("first client = %+v", clients[0])
	}
	if clients[1].State != "preauthenticated" {
		t.Fatalf("second state = %q", clients[1].State)
	}
}

func TestParseSystemBoard_ExtractsOpenWrtBoardJSON(t *testing.T) {
	board := gateway.ParseSystemBoard(`{
	"kernel": "5.15.150",
	"hostname": "astrolink-router",
	"system": "MediaTek MT7621 ver:1 eco:3",
	"model": "GL.iNet GL-MT3000",
	"board_name": "glinet,gl-mt3000",
	"release": {
		"distribution": "OpenWrt",
		"version": "23.05.3",
		"revision": "r23809-234f1a2efa"
	}
}`)

	if board.Hostname != "astrolink-router" {
		t.Fatalf("hostname = %q", board.Hostname)
	}
	if board.Model != "GL.iNet GL-MT3000" {
		t.Fatalf("model = %q", board.Model)
	}
	if board.Firmware != "OpenWrt 23.05.3 r23809-234f1a2efa" {
		t.Fatalf("firmware = %q", board.Firmware)
	}
}

func TestParseOpenNDSLogs_ReturnsRecentNonEmptyLines(t *testing.T) {
	lines := gateway.ParseOpenNDSLogs(`

Thu May 21 10:00:00 2026 daemon.notice opennds[123]: client authenticated
Thu May 21 10:01:00 2026 daemon.warn opennds[123]: client timeout
`, 1)

	if len(lines) != 1 {
		t.Fatalf("len(lines) = %d, want 1", len(lines))
	}
	if lines[0] != "Thu May 21 10:01:00 2026 daemon.warn opennds[123]: client timeout" {
		t.Fatalf("line = %q", lines[0])
	}
}

func TestBuildRouterDiagnostic_ComposesParsedOutputs(t *testing.T) {
	diagnostic := gateway.BuildRouterDiagnostic(
		"Version: 10.2.0\nCurrent clients: 1\n",
		"AA:BB:CC:DD:EE:FF 192.168.1.23 authenticated\n",
		`{"hostname":"router","model":"OpenWrt One","release":{"distribution":"OpenWrt","version":"24.10.0"}}`,
		"line 1\nline 2\n",
	)

	if !diagnostic.Online {
		t.Fatal("expected diagnostic online")
	}
	if diagnostic.OpenNDS.Version != "10.2.0" {
		t.Fatalf("opennds version = %q", diagnostic.OpenNDS.Version)
	}
	if diagnostic.ClientCount != 1 || len(diagnostic.Clients) != 1 {
		t.Fatalf("client counts = %d/%d", diagnostic.ClientCount, len(diagnostic.Clients))
	}
	if diagnostic.Board.Hostname != "router" {
		t.Fatalf("board = %+v", diagnostic.Board)
	}
	if len(diagnostic.RecentLogs) != 2 {
		t.Fatalf("recent logs = %+v", diagnostic.RecentLogs)
	}
}
