package handlers

import (
	"encoding/json"
	"net/http"
	"strings"

	"celldock-web/models"
	"celldock-web/services"
)

func RegisterAPIRoutes(mux *http.ServeMux) {
	svc := services.GetModemService()

	// GET /api/modules
	mux.HandleFunc("/api/modules", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		modules := svc.ListModules()
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 200,
			"data": modules,
		})
	})

	// GET /api/sms, POST /api/sms/send, DELETE /api/sms
	mux.HandleFunc("/api/sms", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodGet {
			messages := svc.ListSMS()
			json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 200,
				"data": messages,
			})
			return
		}

		if r.Method == http.MethodPost {
			var req struct {
				ModuleID string `json:"module_id"`
				Receiver string `json:"receiver"`
				Content  string `json:"content"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
			msg, err := svc.SendSMS(req.ModuleID, req.Receiver, req.Content)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 200,
				"msg":  "短信发送成功",
				"data": msg,
			})
			return
		}

		if r.Method == http.MethodDelete {
			id := r.URL.Query().Get("id")
			ok := svc.DeleteSMS(id)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 200,
				"ok":   ok,
			})
			return
		}
	})

	// GET /api/calls, POST /api/calls/dial, POST /api/calls/hangup
	mux.HandleFunc("/api/calls", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodGet {
			records := svc.ListCallRecords()
			json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 200,
				"data": records,
			})
			return
		}

		if r.Method == http.MethodPost {
			if strings.HasSuffix(r.URL.Path, "/hangup") {
				var req struct {
					ModuleID string `json:"module_id"`
				}
				_ = json.NewDecoder(r.Body).Decode(&req)
				_ = svc.HangupCall(req.ModuleID)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"code": 200,
					"msg":  "挂断信号已发出",
				})
				return
			}

			var req struct {
				ModuleID    string `json:"module_id"`
				PhoneNumber string `json:"phone_number"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
			rec, err := svc.InitiateCall(req.ModuleID, req.PhoneNumber)
			if err != nil {
				http.Error(w, err.Error(), 500)
				return
			}
			json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 200,
				"msg":  "拨号请求成功",
				"data": rec,
			})
			return
		}
	})

	// GET/POST /api/proxy
	mux.HandleFunc("/api/proxy", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodGet {
			modID := r.URL.Query().Get("module_id")
			cfg := svc.GetProxyConfig(modID)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 200,
				"data": cfg,
			})
			return
		}

		if r.Method == http.MethodPost {
			var cfg models.SOCKSProxyConfig
			if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
			svc.SaveProxyConfig(&cfg)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 200,
				"msg":  "SOCKS5 代理配置已保存",
				"data": cfg,
			})
			return
		}
	})

	// GET /api/esim
	mux.HandleFunc("/api/esim", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		modID := r.URL.Query().Get("module_id")
		profiles := svc.ListESIMProfiles(modID)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code": 200,
			"data": profiles,
		})
	})

	// POST /api/at/exec, GET /api/at/logs
	mux.HandleFunc("/api/at", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodGet {
			logs := svc.GetATLogs()
			json.NewEncoder(w).Encode(map[string]interface{}{
				"code": 200,
				"data": logs,
			})
			return
		}

		if r.Method == http.MethodPost {
			var req struct {
				Port    string `json:"port"`
				Command string `json:"command"`
			}
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), 400)
				return
			}
			resp := svc.SendATCommand(req.Port, strings.TrimSpace(req.Command))
			json.NewEncoder(w).Encode(map[string]interface{}{
				"code":     200,
				"response": resp,
			})
			return
		}
	})
}
