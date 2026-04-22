package server

import (
    "context"
    "encoding/json"
    "fmt"
    "os"
    "os/signal"
    "path/filepath"
    "strings"
    "syscall"

    "github.com/xtls/xray-core/core"
    "github.com/xtls/xray-core/infra/conf/serial"

    "go-vless-server/internal/config"
    "go-vless-server/internal/logger"
)

type VLESServer struct {
    config *config.Config
    logger *logger.FilteredLogger
    server core.Server
    cancel context.CancelFunc
}

func NewVLESServer(cfg *config.Config, logger *logger.FilteredLogger) *VLESServer {
    return &VLESServer{
        config: cfg,
        logger: logger,
    }
}

func (s *VLESServer) generateXrayConfig() ([]byte, error) {
    // 1. Подготовка списка клиентов.
    clientsTCP := make([]map[string]interface{}, 0)
    clientsXHTTP := make([]map[string]interface{}, 0)

    for _, client := range s.config.ClientsManager.GetAllClients() {
        if !client.Enabled {
            continue
        }

        // Для TCP/Vision нужен flow
        clientsTCP = append(clientsTCP, map[string]interface{}{
            "id":   client.UUID,
            "flow": "xtls-rprx-vision",
        })

        // Для xHTTP поле flow вообще не нужно, передаем только id
        clientsXHTTP = append(clientsXHTTP, map[string]interface{}{
            "id": client.UUID,
        })
    }

    if len(clientsTCP) == 0 {
        return nil, fmt.Errorf("no active clients found")
    }

    configMap := map[string]interface{}{
        "log": map[string]interface{}{
            "loglevel": "debug",
            "access":   "console",
            "error":    "console",
        },
        "inbounds": []map[string]interface{}{
            // Inbound 1: Стандартный TCP + REALITY (порт 8443)
            {
                "listen":   "0.0.0.0",
                "port":     8443,
                "protocol": "vless",
                "settings": map[string]interface{}{
                    "clients":    clientsTCP,
                    "decryption": "none",
                },
                "streamSettings": map[string]interface{}{
                    "network":  "tcp",
                    "security": "reality",
                    "realitySettings": map[string]interface{}{
                        "dest":        s.config.RealityDest,
                        "serverNames": s.config.RealityServerNames,
                        "privateKey":  s.config.RealityPrivateKey,
                        "shortIds":    s.config.RealityShortIds,
                    },
                    "sniffing": map[string]interface{}{
                        "enabled":      true,
                        "destOverride": []string{"http", "tls"},
                    },
                },
            },
            // Inbound 2: xHTTP + REALITY (порт 8445)
            {
                "listen":   "0.0.0.0",
                "port":     8445,
                "protocol": "vless",
                "settings": map[string]interface{}{
                    "clients":    clientsXHTTP,
                    "decryption": "none",
                },
                "streamSettings": map[string]interface{}{
                    "network":  "xhttp",
                    "security": "reality",
                    "xhttpSettings": map[string]interface{}{
                        "mode": "stream-one",
                        "path": "/s/ref=nav_logo",
			"allowPadding": true,
                    },
                    "realitySettings": map[string]interface{}{
                        "dest":        s.config.RealityDest,
                        "serverNames": s.config.RealityServerNames,
                        "privateKey":  s.config.RealityPrivateKey,
                        "shortIds":    s.config.RealityShortIds,
                    },
                    "sniffing": map[string]interface{}{
                        "enabled":      true,
                        "destOverride": []string{"http", "tls"},
                    },
                },
            },
        },
        "outbounds": []map[string]interface{}{
            {"protocol": "freedom", "tag": "direct"},
            {"protocol": "dns", "tag": "dns-out"},
        },
    }

    configJSON, err := json.MarshalIndent(configMap, "", "  ")
    if err != nil {
        return nil, fmt.Errorf("failed to marshal config: %w", err)
    }

    // Сохраняем конфиг для отладки (очень полезно!)
    configPath := filepath.Join(os.TempDir(), "vless_config.json")
    if err := os.WriteFile(configPath, configJSON, 0644); err != nil {
        s.logger.Warnf("Failed to write config to file: %v", err)
    } else {
        s.logger.Infof("Config saved to %s", configPath)
    }

    return configJSON, nil
}

func (s *VLESServer) Start(ctx context.Context) error {
    s.logger.Info("Generating Xray configuration...")

    // Генерируем конфигурацию
    configJSON, err := s.generateXrayConfig()
    if err != nil {
        return fmt.Errorf("failed to generate config: %w", err)
    }

    // Загружаем конфигурацию Xray-core
    config, err := serial.LoadJSONConfig(strings.NewReader(string(configJSON)))
    if err != nil {
        return fmt.Errorf("failed to load Xray config: %w", err)
    }

    // Создаем экземпляр Xray-core
    server, err := core.New(config)
    if err != nil {
        return fmt.Errorf("failed to create Xray instance: %w", err)
    }
    s.server = server

    // Запускаем сервер
    if err := server.Start(); err != nil {
        return fmt.Errorf("failed to start Xray: %w", err)
    }

    s.logger.Infof("VLESS+REALITY server started")
    s.logger.Infof("  - TCP inbound on port 8443 (for existing clients)")
    s.logger.Infof("Active clients: %d", len(s.config.ClientsManager.GetAllClients()))
    s.logger.Info("Xray is running as embedded library with debug logging")

    // Создаем контекст с отменой для graceful shutdown
    ctx, cancel := context.WithCancel(ctx)
    s.cancel = cancel

    // Обрабатываем сигналы ОС
    go s.handleSignals()

    // Ждем сигнала завершения
    <-ctx.Done()

    return s.Stop()
}

func (s *VLESServer) handleSignals() {
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

    sig := <-sigCh
    s.logger.Infof("Received signal: %v", sig)
    if s.cancel != nil {
        s.cancel()
    }
}

func (s *VLESServer) Stop() error {
    s.logger.Info("Shutting down VLESS server...")

    if s.server != nil {
        if err := s.server.Close(); err != nil {
            s.logger.Errorf("Error closing Xray server: %v", err)
        }
        s.logger.Info("Xray instance closed successfully")
    }

    s.logger.Info("VLESS server stopped successfully")
    return nil
}

// Reload пересоздаёт сервер с новой конфигурацией
func (s *VLESServer) Reload() error {
    s.logger.Info("Reloading Xray configuration...")

    if s.server != nil {
        if err := s.server.Close(); err != nil {
            s.logger.Errorf("Error closing old server: %v", err)
        }
    }

    configJSON, err := s.generateXrayConfig()
    if err != nil {
        return fmt.Errorf("failed to generate new config: %w", err)
    }

    config, err := serial.LoadJSONConfig(strings.NewReader(string(configJSON)))
    if err != nil {
        return fmt.Errorf("failed to load new config: %w", err)
    }

    server, err := core.New(config)
    if err != nil {
        return fmt.Errorf("failed to create new Xray instance: %w", err)
    }

    if err := server.Start(); err != nil {
        return fmt.Errorf("failed to start new Xray instance: %w", err)
    }

    s.server = server
    s.logger.Info("Configuration reloaded successfully")
    return nil
}
