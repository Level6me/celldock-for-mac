package services

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"celldock-web/models"
	"celldock-web/utils"
)

type ModemService struct {
	mu           sync.RWMutex
	modules      map[string]*models.CellularModule
	smsMessages  []*models.SMSMessage
	callRecords  []*models.CallRecord
	proxyConfigs map[string]*models.SOCKSProxyConfig
	esimProfiles map[string][]*models.ESIMProfile
	atLogs       []string
}

var globalModemService *ModemService
var once sync.Once

func GetModemService() *ModemService {
	once.Do(func() {
		globalModemService = &ModemService{
			modules:      make(map[string]*models.CellularModule),
			smsMessages:  make([]*models.SMSMessage, 0),
			callRecords:  make([]*models.CallRecord, 0),
			proxyConfigs: make(map[string]*models.SOCKSProxyConfig),
			esimProfiles: make(map[string][]*models.ESIMProfile),
			atLogs:       make([]string, 0),
		}
		// Async hardware scanning loop
		go globalModemService.startHardwareScanner()
		// Async AT+CLCC call state polling loop
		go globalModemService.startCallStateMonitor()
	})
	return globalModemService
}

func (s *ModemService) startCallStateMonitor() {
	ticker := time.NewTicker(1 * time.Second)
	for range ticker.C {
		s.checkActiveCallState()
	}
}

func (s *ModemService) checkActiveCallState() {
	s.mu.Lock()
	var activeRec *models.CallRecord
	for _, rec := range s.callRecords {
		if rec.Status == "dialing" || rec.Status == "ringing" || rec.Status == "active" {
			activeRec = rec
			break
		}
	}
	s.mu.Unlock()

	if activeRec == nil {
		return
	}

	port := "/dev/ttyUSB2"
	resp, err := utils.ExecATCommand(port, "AT+CLCC\r\n", 400*time.Millisecond)
	if err != nil {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if !strings.Contains(resp, "+CLCC:") {
		if activeRec.Status != "ended" {
			activeRec.Status = "ended"
			activeRec.DurationSec = int(time.Since(activeRec.StartTime).Seconds())
			if activeRec.DurationSec <= 0 {
				activeRec.DurationSec = 1
			}
			s.atLogs = append(s.atLogs, fmt.Sprintf("[%s] 对方已挂断/通话自然结束, 持续时间: %d秒", time.Now().Format("15:04:05"), activeRec.DurationSec))
		}
		return
	}

	// stat: 0=active (已接通), 1=held, 2=dialing (正在拨号), 3=alerting (正在响铃)
	if match := regexp.MustCompile(`\+CLCC:\s*\d+,\s*\d+,\s*(\d+)`).FindStringSubmatch(resp); len(match) > 1 {
		statCode := match[1]
		switch statCode {
		case "2":
			activeRec.Status = "dialing"
		case "3":
			activeRec.Status = "ringing"
		case "0":
			activeRec.Status = "active"
		}
	}
}

func (s *ModemService) startHardwareScanner() {
	// First initial scan
	s.ScanRealHardware()

	// Periodic loop every 5 seconds
	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
		s.ScanRealHardware()
	}
}

func (s *ModemService) ScanRealHardware() {
	ports := []string{"/dev/ttyUSB2", "/dev/ttyUSB3", "/dev/ttyUSB1", "/dev/ttyUSB0"}

	for i, port := range ports {
		if fi, err := os.Stat(port); err != nil || fi.IsDir() {
			continue
		}

		resp, err := utils.ExecATCommand(port, "ATI;+CSQ;+COPS?\r\n", 800*time.Millisecond)
		if err != nil || !strings.Contains(resp, "OK") {
			continue
		}

		id := fmt.Sprintf("mod_hw_%d", i+1)
		modName := "Baiwang QDC507 5G/4G Modem"
		if strings.Contains(resp, "EG25") {
			modName = "Quectel EG25-G Modem"
		}

		signalCSQ := 26
		signalPercent := 83
		if match := regexp.MustCompile(`\+CSQ:\s*(\d+)`).FindStringSubmatch(resp); len(match) > 1 {
			csq, _ := strconv.Atoi(match[1])
			if csq != 99 {
				signalCSQ = csq
				signalPercent = int(float64(csq) / 31.0 * 100.0)
			}
		}

		operator := "CHN-UNICOM 5G/4G"
		networkTech := "5G/4G LTE"
		if match := regexp.MustCompile(`"([^"]+)"`).FindStringSubmatch(resp); len(match) > 1 {
			operator = match[1]
		}

		s.mu.Lock()
		s.modules[id] = &models.CellularModule{
			ID:            id,
			Name:          modName,
			Port:          port,
			Operator:      operator,
			NetworkTech:   networkTech,
			SignalPercent: signalPercent,
			SignalCSQ:     signalCSQ,
			IPAddress:     "10.0.0.2",
			Status:        "online",
			DownloadBps:   14200000,
			UploadBps:     3800000,
			IsPriority:    true,
			LastSeen:      time.Now(),
		}
		s.mu.Unlock()
		fmt.Printf("[ModemService] Real Cellular Modem Discovered on %s (%s, Operator: %s, CSQ: %d)\n", port, modName, operator, signalCSQ)
		break
	}
}

func (s *ModemService) ListModules() []*models.CellularModule {
	s.mu.RLock()
	defer s.mu.RUnlock()

	res := make([]*models.CellularModule, 0, len(s.modules))
	for _, m := range s.modules {
		res = append(res, m)
	}
	return res
}

func (s *ModemService) GetModule(id string) (*models.CellularModule, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	m, ok := s.modules[id]
	if !ok {
		return nil, fmt.Errorf("modem %s not found", id)
	}
	return m, nil
}

func (s *ModemService) ListSMS() []*models.SMSMessage {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.smsMessages
}

func (s *ModemService) SendSMS(moduleID, receiver, content string) (*models.SMSMessage, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	mod, ok := s.modules[moduleID]
	port := "/dev/ttyUSB2"
	if ok {
		port = mod.Port
	}

	go func() {
		_, _ = utils.ExecATCommand(port, "AT+CMGF=1", 1*time.Second)
		sendCmd := fmt.Sprintf("AT+CMGS=\"%s\"", receiver)
		_, _ = utils.ExecATCommand(port, sendCmd, 1*time.Second)
		rawResp, _ := utils.ExecATCommand(port, content+"\x1A", 5*time.Second)
		s.mu.Lock()
		s.atLogs = append(s.atLogs, fmt.Sprintf("[%s] %s -> Send SMS to %s: %s", time.Now().Format("15:04:05"), port, receiver, rawResp))
		s.mu.Unlock()
	}()

	msg := &models.SMSMessage{
		ID:             fmt.Sprintf("sms_%d", time.Now().UnixNano()),
		ModuleID:       moduleID,
		Sender:         "本机",
		Receiver:       receiver,
		Content:        content,
		Timestamp:      time.Now(),
		IsRead:         true,
		IsVerification: false,
	}
	s.smsMessages = append([]*models.SMSMessage{msg}, s.smsMessages...)

	return msg, nil
}

func (s *ModemService) DeleteSMS(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, msg := range s.smsMessages {
		if msg.ID == id {
			s.smsMessages = append(s.smsMessages[:i], s.smsMessages[i+1:]...)
			return true
		}
	}
	return false
}

func (s *ModemService) ListCallRecords() []*models.CallRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.callRecords
}

func (s *ModemService) InitiateCall(moduleID, phoneNumber string) (*models.CallRecord, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	mod, ok := s.modules[moduleID]
	port := "/dev/ttyUSB2"
	if ok && mod.Port != "" {
		port = mod.Port
	}

	go func() {
		dialCmd := fmt.Sprintf("ATD%s;", phoneNumber)
		rawResp, _ := utils.ExecATCommand(port, dialCmd, 2*time.Second)
		time.Sleep(500 * time.Millisecond)
		// Unmute modem mic and set call volume
		_, _ = utils.ExecATCommand(port, "AT+CMUT=0", 500*time.Millisecond)
		_, _ = utils.ExecATCommand(port, "AT+CLVL=5", 500*time.Millisecond)
		s.mu.Lock()
		s.atLogs = append(s.atLogs, fmt.Sprintf("[%s] %s -> Dial %s: %s", time.Now().Format("15:04:05"), port, phoneNumber, rawResp))
		s.mu.Unlock()
	}()

	rec := &models.CallRecord{
		ID:          fmt.Sprintf("call_%d", time.Now().UnixNano()),
		ModuleID:    moduleID,
		PhoneNumber: phoneNumber,
		Direction:   "outbound",
		Status:      "active",
		StartTime:   time.Now(),
		DurationSec: 0,
	}
	s.callRecords = append([]*models.CallRecord{rec}, s.callRecords...)

	return rec, nil
}

func (s *ModemService) HangupCall(moduleID string) error {
	s.mu.Lock()
	// Immediately mark all active calls as ended in memory
	for _, rec := range s.callRecords {
		if rec.Status == "active" {
			rec.Status = "ended"
			rec.DurationSec = int(time.Since(rec.StartTime).Seconds())
			if rec.DurationSec <= 0 {
				rec.DurationSec = 1
			}
		}
	}

	mod, ok := s.modules[moduleID]
	port := "/dev/ttyUSB2"
	if ok && mod.Port != "" {
		port = mod.Port
	}
	s.mu.Unlock()

	go func() {
		rawResp, _ := utils.ExecATCommand(port, "ATH", 2*time.Second)
		s.mu.Lock()
		s.atLogs = append(s.atLogs, fmt.Sprintf("[%s] %s -> Hangup ATH: %s", time.Now().Format("15:04:05"), port, rawResp))
		s.mu.Unlock()
	}()

	return nil
}

func (s *ModemService) SendATCommand(port, cmd string) string {
	if port == "" {
		port = "/dev/ttyUSB2"
	}

	rawResp, err := utils.ExecATCommand(port, cmd, 3*time.Second)
	if err != nil {
		rawResp = fmt.Sprintf("硬件串口通信失败: %v", err)
	}

	s.mu.Lock()
	logLine := fmt.Sprintf("[%s] %s -> %s\n%s", time.Now().Format("15:04:05"), port, cmd, rawResp)
	s.atLogs = append(s.atLogs, logLine)
	s.mu.Unlock()

	return rawResp
}

func (s *ModemService) GetATLogs() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.atLogs
}

func (s *ModemService) GetProxyConfig(moduleID string) *models.SOCKSProxyConfig {
	s.mu.RLock()
	defer s.mu.RUnlock()
	cfg, ok := s.proxyConfigs[moduleID]
	if !ok {
		return &models.SOCKSProxyConfig{
			ModuleID:  moduleID,
			Port:      1080,
			AllowLAN:  true,
			IsRunning: true,
		}
	}
	return cfg
}

func (s *ModemService) SaveProxyConfig(cfg *models.SOCKSProxyConfig) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.proxyConfigs[cfg.ModuleID] = cfg
}

func (s *ModemService) ListESIMProfiles(moduleID string) []*models.ESIMProfile {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.esimProfiles[moduleID]
}
