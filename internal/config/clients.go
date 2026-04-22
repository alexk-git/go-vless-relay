package config

import (
    "encoding/json"
    "fmt"
    "os"
    "sync"
    "time"

    "github.com/google/uuid"
)

type Client struct {
    ID        string `json:"id"`
    Name      string `json:"name"`
    Email     string `json:"email"`
    UUID      string `json:"uuid"`
    CreatedAt string `json:"created_at"`
    Enabled   bool   `json:"enabled"`
}

type ClientsManager struct {
    mu       sync.RWMutex
    clients  map[string]*Client
    filePath string
}

func NewClientsManager(filePath string) (*ClientsManager, error) {
    cm := &ClientsManager{
        clients:  make(map[string]*Client),
        filePath: filePath,
    }
    if err := cm.load(); err != nil {
        return nil, err
    }
    return cm, nil
}

func (cm *ClientsManager) load() error {
    data, err := os.ReadFile(cm.filePath)
    if err != nil {
        if os.IsNotExist(err) {
            return cm.save()
        }
        return err
    }

    var clientsList []*Client
    if err := json.Unmarshal(data, &clientsList); err != nil {
        return err
    }

    for _, c := range clientsList {
        cm.clients[c.UUID] = c
    }
    return nil
}

func (cm *ClientsManager) save() error {
    clientsList := make([]*Client, 0, len(cm.clients))
    for _, c := range cm.clients {
        clientsList = append(clientsList, c)
    }
    data, err := json.MarshalIndent(clientsList, "", "  ")
    if err != nil {
        return err
    }
    return os.WriteFile(cm.filePath, data, 0644)
}

func (cm *ClientsManager) AddClient(name, email string) (*Client, error) {
    cm.mu.Lock()
    defer cm.mu.Unlock()

    newUUID := uuid.New().String()
    client := &Client{
        ID:        uuid.New().String(),
        Name:      name,
        Email:     email,
        UUID:      newUUID,
        CreatedAt: time.Now().Format(time.RFC3339),
        Enabled:   true,
    }

    cm.clients[newUUID] = client
    if err := cm.save(); err != nil {
        return nil, err
    }
    return client, nil
}

func (cm *ClientsManager) GetClientByUUID(uuid string) (*Client, bool) {
    cm.mu.RLock()
    defer cm.mu.RUnlock()

    client, exists := cm.clients[uuid]
    if !exists {
        return nil, false
    }
    return client, client.Enabled
}

func (cm *ClientsManager) GetAllClients() []*Client {
    cm.mu.RLock()
    defer cm.mu.RUnlock()

    clients := make([]*Client, 0, len(cm.clients))
    for _, c := range cm.clients {
        clients = append(clients, c)
    }
    return clients
}

func (cm *ClientsManager) RemoveClient(uuid string) error {
    cm.mu.Lock()
    defer cm.mu.Unlock()

    if _, exists := cm.clients[uuid]; !exists {
        return fmt.Errorf("client with UUID %s not found", uuid)
    }

    delete(cm.clients, uuid)
    return cm.save()
}

func (cm *ClientsManager) DisableClient(uuid string) error {
    cm.mu.Lock()
    defer cm.mu.Unlock()

    client, exists := cm.clients[uuid]
    if !exists {
        return fmt.Errorf("client with UUID %s not found", uuid)
    }

    client.Enabled = false
    return cm.save()
}

func (cm *ClientsManager) EnableClient(uuid string) error {
    cm.mu.Lock()
    defer cm.mu.Unlock()

    client, exists := cm.clients[uuid]
    if !exists {
        return fmt.Errorf("client with UUID %s not found", uuid)
    }

    client.Enabled = true
    return cm.save()
}
