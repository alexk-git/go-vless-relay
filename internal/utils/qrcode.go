package utils

import (
    "fmt"

    qrcode "github.com/skip2/go-qrcode"
)

func GenerateQRCode(content, outputPath string) (string, error) {
    if outputPath == "" {
        outputPath = fmt.Sprintf("/tmp/qrcode_%p.png", &content)
    }

    err := qrcode.WriteFile(content, qrcode.Medium, 300, outputPath)
    if err != nil {
        return "", err
    }

    return outputPath, nil
}

func GenerateQRCodeToBytes(content string) ([]byte, error) {
    return qrcode.Encode(content, qrcode.Medium, 300)
}
