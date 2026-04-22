package config

import (
    "bufio"
    "fmt"
    "os"
    "strconv"
    "strings"
)

func LoadConfig(configPath string) (*Config, error) {
    props, err := readPropertiesFile(configPath)
    if err != nil {
        return nil, fmt.Errorf("ошибка чтения конфигурации: %w", err)
    }

    port, _ := strconv.Atoi(getProp(props, "SERVER_PORT", "443"))
    debugMode := getProp(props, "DEBUG_MODE", "false") == "true"
    logLevel := getProp(props, "LOG_LEVEL", "info")

    serverNames := parseStringList(getProp(props, "REALITY_SERVER_NAMES", `["www.google.com"]`))
    shortIds := parseStringList(getProp(props, "REALITY_SHORT_IDS", `["12345678"]`))
    dnsServers := parseStringList(getProp(props, "DNS_SERVERS", `["8.8.8.8", "1.1.1.1"]`))

    clientsDBPath := getProp(props, "CLIENTS_DB_PATH", "data/clients.json")
    clientsManager, err := NewClientsManager(clientsDBPath)
    if err != nil {
        return nil, fmt.Errorf("ошибка загрузки клиентов: %w", err)
    }

    return &Config{
        Port:               port,
        ServerAddr:         getProp(props, "SERVER_ADDR", "localhost"),
        RealityDest:        getProp(props, "REALITY_DEST", "www.google.com:443"),
        RealityServerNames: serverNames,
        RealityPrivateKey:  getProp(props, "REALITY_PRIVATE_KEY", ""),
        RealityPublicKey:   getProp(props, "REALITY_PUBLIC_KEY", ""),
        RealityShortIds:    shortIds,
        DNSServers:         dnsServers,
        DebugMode:          debugMode,
        LogLevel:           logLevel,
        ClientsManager:     clientsManager,
    }, nil
}

func readPropertiesFile(filename string) (map[string]string, error) {
    file, err := os.Open(filename)
    if err != nil {
        return nil, err
    }
    defer file.Close()

    properties := make(map[string]string)
    scanner := bufio.NewScanner(file)

    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())
        if line == "" || strings.HasPrefix(line, "#") {
            continue
        }
        parts := strings.SplitN(line, "=", 2)
        if len(parts) != 2 {
            continue
        }
        properties[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
    }

    return properties, scanner.Err()
}

func getProp(props map[string]string, key, defaultValue string) string {
    if val, ok := props[key]; ok && val != "" {
        return val
    }
    return defaultValue
}

func parseStringList(s string) []string {
    s = strings.Trim(s, "[]")
    if s == "" {
        return []string{}
    }
    parts := strings.Split(s, ",")
    result := make([]string, len(parts))
    for i, p := range parts {
        result[i] = strings.Trim(strings.TrimSpace(p), `"`)
    }
    return result
}
