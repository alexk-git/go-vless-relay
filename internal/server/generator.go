package server

import (
    "crypto/rand"
    "encoding/base64"
    "fmt"

    "golang.org/x/crypto/curve25519"
)

// generateX25519KeyPair генерирует пару ключей X25519 в формате, совместимом с Xray.
// Это точная реплика того, как работает команда ./xray x25519.
func generateX25519KeyPair() (privateKey, publicKey string, err error) {
    // 1. Генерируем 32 случайных байта для приватного ключа
    privateKeyBytes := make([]byte, curve25519.ScalarSize)
    if _, err = rand.Read(privateKeyBytes); err != nil {
        return "", "", fmt.Errorf("failed to generate private key: %w", err)
    }

    // 2. Применяем "clamping" (зажим) согласно спецификации X25519
    //    Это та самая критически важная операция, которая отличает корректный ключ от некорректного.
    privateKeyBytes[0] &= 248
    privateKeyBytes[31] &= 127
    privateKeyBytes[31] |= 64

    // 3. Вычисляем публичный ключ на основе приватного
    publicKeyBytes, err := curve25519.X25519(privateKeyBytes, curve25519.Basepoint)
    if err != nil {
        return "", "", fmt.Errorf("failed to derive public key: %w", err)
    }

    // 4. Кодируем оба ключа в Base64RawURL (без padding, с символами '-' и '_')
    return base64.RawURLEncoding.EncodeToString(privateKeyBytes),
        base64.RawURLEncoding.EncodeToString(publicKeyBytes),
        nil
}

// GenerateKeyPair - публичная обертка для функции generateX25519KeyPair.
func GenerateKeyPair() (privateKey, publicKey string, err error) {
    return generateX25519KeyPair()
}

// GenerateShortID генерирует случайный короткий идентификатор (до 16 hex-символов).
func GenerateShortID() string {
    id := make([]byte, 8) // 8 байт = 16 hex символов
    rand.Read(id)
    return fmt.Sprintf("%016x", id) // Всегда 16 символов, дополняя нулями слева
}

// GenerateShortIDs генерирует несколько shortId.
func GenerateShortIDs(count int) []string {
    ids := make([]string, count)
    for i := 0; i < count; i++ {
        ids[i] = GenerateShortID()
    }
    return ids
}
