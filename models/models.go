package models

import "time"

// CellularModule represents a connected USB cellular modem (e.g. QDC507).
type CellularModule struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Port          string    `json:"port"`
	Operator      string    `json:"operator"`
	NetworkTech   string    `json:"network_tech"` // e.g. "5G SA", "LTE", "3G"
	SignalPercent int       `json:"signal_percent"`
	SignalCSQ     int       `json:"signal_csq"`
	IPAddress     string    `json:"ip_address"`
	Status        string    `json:"status"` // "online", "connecting", "offline"
	DownloadBps   int64     `json:"download_bps"`
	UploadBps     int64     `json:"upload_bps"`
	IsPriority    bool      `json:"is_priority"`
	LastSeen      time.Time `json:"last_seen"`
}

// SMSMessage represents a received or sent text message.
type SMSMessage struct {
	ID               string    `json:"id"`
	ModuleID         string    `json:"module_id"`
	Sender           string    `json:"sender"`
	Receiver         string    `json:"receiver"`
	Content          string    `json:"content"`
	Timestamp        time.Time `json:"timestamp"`
	IsRead           bool      `json:"is_read"`
	IsVerification   bool      `json:"is_verification"`
	VerificationCode string    `json:"verification_code,omitempty"`
}

// CallRecord represents a voice call log entry.
type CallRecord struct {
	ID           string    `json:"id"`
	ModuleID     string    `json:"module_id"`
	PhoneNumber  string    `json:"phone_number"`
	Direction    string    `json:"direction"` // "inbound", "outbound"
	Status       string    `json:"status"`    // "ringing", "active", "ended", "missed"
	StartTime    time.Time `json:"start_time"`
	DurationSec  int       `json:"duration_sec"`
	RecordPath   string    `json:"record_path,omitempty"`
}

// SOCKSProxyConfig represents SOCKS5 proxy settings for a module.
type SOCKSProxyConfig struct {
	ModuleID     string `json:"module_id"`
	Port         int    `json:"port"`
	AllowLAN     bool   `json:"allow_lan"`
	AuthRequired bool   `json:"auth_required"`
	Username     string `json:"username"`
	Password     string `json:"password"`
	IsRunning    bool   `json:"is_running"`
}

// ESIMProfile represents an eUICC / eSIM profile entry.
type ESIMProfile struct {
	ICCID       string `json:"iccid"`
	ProfileName string `json:"profile_name"`
	Provider    string `json:"provider"`
	IsEnabled   bool   `json:"is_enabled"`
	State       string `json:"state"` // "enabled", "disabled"
}
