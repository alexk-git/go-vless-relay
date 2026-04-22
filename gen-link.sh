#!/bin/bash

# Читаем UUID из clients.json
UUID=$(cat data/clients.json | grep -o '"uuid":"[^"]*"' | head -1 | cut -d'"' -f4)

# Читаем настройки из .env
SERVER_ADDR=$(grep "^SERVER_ADDR=" data/.env | cut -d'=' -f2)
PORT=$(grep "^SERVER_PORT=" data/.env | cut -d'=' -f2)
PUBLIC_KEY=$(grep "^REALITY_PUBLIC_KEY=" data/.env | cut -d'=' -f2)
SNI=$(grep "^REALITY_SERVER_NAMES=" data/.env | cut -d'=' -f2 | sed 's/\["\([^"]*\)".*/\1/')
SHORT_ID=$(grep "^REALITY_SHORT_IDS=" data/.env | cut -d'=' -f2 | sed 's/\["\([^"]*\)".*/\1/')

# Генерируем ссылку
echo "vless://$UUID@$SERVER_ADDR:$PORT?security=reality&sni=$SNI&fp=chrome&pbk=$PUBLIC_KEY&sid=$SHORT_ID#Admin"
