package main

import (
	"fmt"
	"log"
	"net/http"

	"celldock-web/config"
	"celldock-web/frontend"
	"celldock-web/handlers"
	"celldock-web/services"
)

func main() {
	cfg := config.LoadConfig()

	// Initialize Modem Service
	_ = services.GetModemService()

	mux := http.NewServeMux()

	// Register API & Static WebUI Routes
	handlers.RegisterAPIRoutes(mux)
	frontend.RegisterStaticRoutes(mux)

	fmt.Printf("====================================================\n")
	fmt.Printf("🚀 CellDock Web 守护进程与控制台已启动！\n")
	fmt.Printf("🌐 运行监听地址: http://127.0.0.1%s\n", cfg.ListenAddr)
	fmt.Printf("📂 数据持久化路径: %s\n", cfg.DataDir)
	fmt.Printf("====================================================\n")

	if err := http.ListenAndServe(cfg.ListenAddr, mux); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
