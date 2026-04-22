package main

import (
    "context"
    "flag"
    "fmt"
    "log"
    "os"
    "os/signal"
    "strings"
    "syscall"

    "go-vless-server/internal/config"
    "go-vless-server/internal/logger"
    "go-vless-server/internal/server"
    "go-vless-server/internal/utils"

    // Blank imports для регистрации компонентов Xray-core
    _ "github.com/xtls/xray-core/app/dispatcher"
    _ "github.com/xtls/xray-core/app/dns"
    _ "github.com/xtls/xray-core/app/log"
    _ "github.com/xtls/xray-core/app/proxyman/inbound"
    _ "github.com/xtls/xray-core/app/proxyman/outbound"
    _ "github.com/xtls/xray-core/app/router"
    _ "github.com/xtls/xray-core/app/stats"
    _ "github.com/xtls/xray-core/proxy/blackhole"
    _ "github.com/xtls/xray-core/proxy/dns"
    _ "github.com/xtls/xray-core/proxy/freedom"
    _ "github.com/xtls/xray-core/proxy/http"
    _ "github.com/xtls/xray-core/proxy/socks"
    _ "github.com/xtls/xray-core/proxy/trojan"
    _ "github.com/xtls/xray-core/proxy/vless"
    _ "github.com/xtls/xray-core/proxy/vmess"
    _ "github.com/xtls/xray-core/transport/internet/kcp"
    _ "github.com/xtls/xray-core/transport/internet/tcp"
    _ "github.com/xtls/xray-core/transport/internet/tls"
    _ "github.com/xtls/xray-core/transport/internet/websocket"
    _ "github.com/xtls/xray-core/transport/internet/reality"
    _ "github.com/xtls/xray-core/transport/internet/splithttp"
)

func main() {
    var (
        configPath string
        showHelp   bool
        addUser    string
        addEmail   string
        listUsers  bool
        removeUser string
        debugMode  bool
    )

    flag.StringVar(&configPath, "config", "data/.env", "Path to config file")
    flag.BoolVar(&showHelp, "help", false, "Show help")
    flag.StringVar(&addUser, "add-user", "", "Add new user with given name")
    flag.StringVar(&addEmail, "email", "", "Email for new user")
    flag.BoolVar(&listUsers, "list-users", false, "List all users")
    flag.StringVar(&removeUser, "remove-user", "", "Remove user by UUID")
    flag.BoolVar(&debugMode, "debug", false, "Enable debug mode")
    flag.Parse()

    if showHelp {
        printHelp()
        return
    }

    cfg, err := config.LoadConfig(configPath)
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    if debugMode {
        cfg.DebugMode = true
        cfg.LogLevel = "debug"
        log.Println("🔍 Debug mode enabled")
    }

    if cfg.RealityPrivateKey == "" || cfg.RealityPublicKey == "" {
        log.Println("REALITY keys not found, generating new ones...")
        privateKey, publicKey, err := server.GenerateKeyPair()
        if err != nil {
            log.Fatalf("Failed to generate REALITY keys: %v", err)
        }

        if err := saveKeysToEnv(configPath, privateKey, publicKey); err != nil {
            log.Printf("Warning: could not save keys to .env: %v", err)
        }

        cfg.RealityPrivateKey = privateKey
        cfg.RealityPublicKey = publicKey
        log.Println("REALITY keys generated successfully")
    }

    logger := logger.NewFilteredLogger(cfg.DebugMode, cfg.LogLevel)

    if addUser != "" {
        email := addEmail
        if email == "" {
            email = fmt.Sprintf("%s@localhost", addUser)
        }
        client, err := cfg.ClientsManager.AddClient(addUser, email)
        if err != nil {
            log.Fatalf("Failed to add user: %v", err)
        }
        fmt.Printf("User added:\n")
        fmt.Printf("  UUID: %s\n", client.UUID)
        fmt.Printf("  Name: %s\n", client.Name)
        fmt.Printf("  Email: %s\n", client.Email)
        return
    }

    if listUsers {
        clients := cfg.ClientsManager.GetAllClients()
        fmt.Printf("Total clients: %d\n", len(clients))
        for _, c := range clients {
            fmt.Printf("  %s | %s | %s | Enabled: %v\n", c.UUID, c.Name, c.Email, c.Enabled)
        }
        return
    }

    if removeUser != "" {
        if err := cfg.ClientsManager.RemoveClient(removeUser); err != nil {
            log.Fatalf("Failed to remove user: %v", err)
        }
        fmt.Printf("User %s removed\n", removeUser)
        return
    }

    // ============================================
    // Вывод ссылок для клиентов
    // ============================================
    clients := cfg.ClientsManager.GetAllClients()
    for _, client := range clients {
        if !client.Enabled {
            continue
        }

        // TCP ссылка (порт 8443) - для существующих клиентов
        tcpLink := fmt.Sprintf("vless://%s@%s:8443?encryption=none&flow=xtls-rprx-vision&security=reality&sni=%s&fp=chrome&pbk=%s&sid=%s&type=tcp#%s_TCP",
            client.UUID,
            cfg.ServerAddr,
            cfg.GetMainSNI(),
            cfg.RealityPublicKey,
            cfg.GetMainShortID(),
            client.Name,
        )

        // xHTTP ссылка (порт 8445) - для новых клиентов
        xhttpLink := fmt.Sprintf("vless://%s@%s:8445?encryption=none&security=reality&sni=%s&fp=chrome&pbk=%s&sid=%s&type=xhttp&path=/s/ref=nav_logo&mode=stream-one#%s_XHTTP",
            client.UUID,
            cfg.ServerAddr,
            cfg.GetMainSNI(),
            cfg.RealityPublicKey,
            cfg.GetMainShortID(),
            client.Name,
        )

        logger.Infof("=== Client: %s ===", client.Name)
        logger.Infof("TCP link (port 8443): %s", tcpLink)
        logger.Infof("XHTTP link (port 8445): %s", xhttpLink)

        qrPath, _ := utils.GenerateQRCode(tcpLink, fmt.Sprintf("/tmp/vless_%s_tcp.png", client.ID))
        logger.Infof("TCP QR code saved: %s", qrPath)

        qrPathX, _ := utils.GenerateQRCode(xhttpLink, fmt.Sprintf("/tmp/vless_%s_xhttp.png", client.ID))
        logger.Infof("XHTTP QR code saved: %s", qrPathX)
    }

    vlessServer := server.NewVLESServer(cfg, logger)

    ctx, cancel := signal.NotifyContext(context.Background(),
        syscall.SIGINT, syscall.SIGTERM)
    defer cancel()

    if err := vlessServer.Start(ctx); err != nil {
        log.Fatalf("Failed to start server: %v", err)
    }

    logger.Info("Server stopped")
}

func printHelp() {
    fmt.Println(`VLESS+REALITY Server

Usage:
  ./vless-server [options]

Options:
  -config <path>      Path to config file (default: data/.env)
  -debug              Enable debug mode
  -help               Show this help message

User management:
  -add-user <name>    Add new user
  -email <email>      Email for new user (optional)
  -list-users         List all users
  -remove-user <uuid> Remove user by UUID

Examples:
  ./vless-server
  ./vless-server -add-user "Alice" -email "alice@example.com"
  ./vless-server -list-users
  ./vless-server -remove-user "a1b2c3d4-..."`)
}

func saveKeysToEnv(configPath, privateKey, publicKey string) error {
    content, err := os.ReadFile(configPath)
    if err != nil {
        return err
    }

    lines := strings.Split(string(content), "\n")
    privateKeyFound := false
    publicKeyFound := false

    for i, line := range lines {
        if strings.HasPrefix(line, "REALITY_PRIVATE_KEY=") {
            lines[i] = "REALITY_PRIVATE_KEY=" + privateKey
            privateKeyFound = true
        }
        if strings.HasPrefix(line, "REALITY_PUBLIC_KEY=") {
            lines[i] = "REALITY_PUBLIC_KEY=" + publicKey
            publicKeyFound = true
        }
    }

    if !privateKeyFound {
        lines = append(lines, "REALITY_PRIVATE_KEY="+privateKey)
    }
    if !publicKeyFound {
        lines = append(lines, "REALITY_PUBLIC_KEY="+publicKey)
    }

    return os.WriteFile(configPath, []byte(strings.Join(lines, "\n")), 0644)
}
