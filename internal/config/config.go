package config

type Config struct {
    Port               int      `json:"port"`
    ServerAddr         string   `json:"server_addr"`
    RealityDest        string   `json:"reality_dest"`
    RealityServerNames []string `json:"reality_server_names"`
    RealityPrivateKey  string   `json:"reality_private_key"`
    RealityPublicKey   string   `json:"reality_public_key"`
    RealityShortIds    []string `json:"reality_short_ids"`
    DNSServers         []string `json:"dns_servers"`
    DebugMode          bool     `json:"debug_mode"`
    LogLevel           string   `json:"log_level"`

    ClientsManager *ClientsManager `json:"-"`
}

func (c *Config) GetMainSNI() string {
    if len(c.RealityServerNames) > 0 {
        return c.RealityServerNames[0]
    }
    return "www.google.com"
}

func (c *Config) GetMainShortID() string {
    if len(c.RealityShortIds) > 0 {
        return c.RealityShortIds[0]
    }
    return "12345678"
}
