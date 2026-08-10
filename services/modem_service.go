package services

import (
	"fmt"
	"sync"
	"time"

	"celldock-web/models"
	"celldock-web/utils"
)

type ModemService struct {
	mu            sync.RWMutex
	modules       map[string]*models.CellularModule
	smsMessages   []*models.SMSMessage
	callRecords   []*models.CallRecord
	proxyConfigs  map[string]*models.SOCKSProxyConfig
	esimProfiles  map[string][]*models.ESIMProfile
	atLogs        []string
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
		globalModemService.initMockData()
	})
	return globalModemService
}

func (s *ModemService) initMockData() {
	// Initialize default simulated/detected modules if available
	mod1 := &models.CellularModule{
		ID:            "mod_qdc507_01",
		Name:          "QDC507 5G Modem #1",
		Port:          "/dev/ttyUSB2",
		Operator:      "China Mobile 5G",
		NetworkTech:   "5G SA",
		SignalPercent: 92,
		SignalCSQ:     29,
		IPAddress:     "10.142.88.102",
		Status:        "online",
		DownloadBps:   15420000,
		UploadBps:     3200000,
		IsPriority:    true,
		LastSeen:      time.Now(),
	}
	mod2 := &models.CellularModule{
		ID:            "mod_qdc507_02",
		Name:          "QDC507 4G Modem #2",
		Port:          "/dev/ttyUSB4",
		Operator:      "China Unicom 4G",
		NetworkTech:   "LTE",
		SignalPercent: 78,
		SignalCSQ:     24,
		IPAddress:     "10.88.19.45",
		Status:        "online",
		DownloadBps:   4800000,
		UploadBps:     1100000,
		IsPriority:    false,
		LastSeen:      time.Now(),
	}
	s.modules[mod1.ID] = mod1
	s.modules[mod2.ID] = mod2

	// Sample SMS Messages
	s.smsMessages = append(s.smsMessages, &models.SMSMessage{
		ID:               "sms_101",
		ModuleID:         mod1.ID,
		Sender:           "10086",
		Receiver:         "13800138000",
		Content:          "【中国移动】尊敬的客户，您的验证码为：849201，有效期30分钟，请勿泄露给他人。",
		Timestamp:        time.Now().Add(-5 * time.Minute),
		IsRead:           false,
		IsVerification:   true,
		VerificationCode: "849201",
	})
	s.smsMessages = append(s.smsMessages, &models.SMSMessage{
		ID:               "sms_102",
		ModuleID:         mod1.ID,
		Sender:           "1069012345",
		Receiver:         "13800138000",
		Content:          "【腾讯云】您的登录验证码是 392150，打死也不要告诉别人哦！",
		Timestamp:        time.Now().Add(-25 * time.Minute),
		IsRead:           true,
		IsVerification:   true,
		VerificationCode: "392150",
	})

	// Sample Call Records
	s.callRecords = append(s.callRecords, &models.CallRecord{
		ID:          "call_201",
		ModuleID:    mod1.ID,
		PhoneNumber: "13800138000",
		Direction:   "inbound",
		Status:      "ended",
		StartTime:   time.Now().Add(-1 * time.Hour),
		DurationSec: 142,
		RecordPath:  "/data/recordings/call_201.wav",
	})

	// Sample Proxy Config
	s.proxyConfigs[mod1.ID] = &models.SOCKSProxyConfig{
		ModuleID:     mod1.ID,
		Port:         1080,
		AllowLAN:     true,
		AuthRequired: false,
		Username:     "admin",
		Password:     "123456",
		IsRunning:    true,
	}

	// Sample eSIM Profiles
	s.esimProfiles[mod1.ID] = []*models.ESIMProfile{
		{
			ICCID:       "89860012345678901234",
			ProfileName: "中国移动 5G 常用卡",
			Provider:    "China Mobile",
			IsEnabled:   true,
			State:       "enabled",
		},
		{
			ICCID:       "89860098765432109876",
			ProfileName: "香港 CSL 漫游卡",
			Provider:    "CSL HK",
			IsEnabled:   false,
			State:       "disabled",
		},
	}
}

func (s *ModemService) ListModules() []*models.CellularModule {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Try scanning real hardware USB serial ports
	ports, _ := utils.ListSerialPorts()
	if len(ports) > 0 {
		for i, p := range ports {
			id := fmt.Sprintf("mod_hw_%d", i+1)
			if _, exists := s.modules[id]; !exists {
				s.modules[id] = &models.CellularModule{
					ID:            id,
					Name:          fmt.Sprintf("USB Cellular Module (%s)", p),
					Port:          p,
					Operator:      "检测中...",
					NetworkTech:   "LTE/5G",
					SignalPercent: 85,
					SignalCSQ:     27,
					IPAddress:     "192.168.8.100",
					Status:        "online",
					LastSeen:      time.Now(),
				}
			}
		}
	}

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

func (s *ModemService) SendATCommand(port, cmd string) string {
	s.mu.Lock()
	defer s.mu.Unlock()

	logLine := fmt.Sprintf("[%s] %s -> %s", time.Now().Format("15:04:05"), port, cmd)
	s.atLogs = append(s.atLogs, logLine)

	// Simulated AT Command responses
	switch cmd {
	case "AT":
		return "OK"
	case "AT+CSQ":
		return "+CSQ: 28,99\r\n\r\nOK"
	case "AT+COPS?":
		return "+COPS: 0,0,\"CHINA MOBILE\",7\r\n\r\nOK"
	case "AT+CPIN?":
		return "+CPIN: READY\r\n\r\nOK"
	case "ATI":
		return "Quectel / QDC507\r\nRevision: QDC507R01A01\r\n\r\nOK"
	default:
		return "OK"
	}
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
			IsRunning: false,
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
