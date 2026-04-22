#!/bin/bash

CLIENTS_FILE="data/clients.json"
ENV_FILE="configs/env.properties"

# Проверка наличия файлов
if [ ! -f "$CLIENTS_FILE" ]; then
    echo "❌ $CLIENTS_FILE not found"
    exit 1
fi

if [ ! -f "$ENV_FILE" ]; then
    echo "❌ $ENV_FILE not found"
    exit 1
fi

# Читаем UUID первого клиента
UUID=$(grep -o '"uuid":"[^"]*"' "$CLIENTS_FILE" | head -1 | cut -d'"' -f4)

if [ -z "$UUID" ]; then
    echo "❌ UUID not found in $CLIENTS_FILE"
    exit 1
fi

# Читаем настройки из .env
source "$ENV_FILE"

# Извлекаем SNI (первый элемент массива)
SNI=$(echo "$REALITY_SERVER_NAMES" | sed 's/\["\([^"]*\)".*/\1/')

# Извлекаем ShortID
SHORT_ID=$(echo "$REALITY_SHORT_IDS" | sed 's/\["\([^"]*\)".*/\1/')

# ============================================
# Ссылка 1: TCP (порт 8443) - для старых клиентов
# ============================================
TCP_LINK="vless://$UUID@$SERVER_ADDR:8443?encryption=none&flow=xtls-rprx-vision&security=reality&sni=$SNI&fp=chrome&pbk=$REALITY_PUBLIC_KEY&sid=$SHORT_ID&type=tcp#Admin_TCP"

# ============================================
# Ссылка 2: xHTTP (порт 8445) - для новых клиентов
# ============================================
XHTTP_LINK="vless://$UUID@$SERVER_ADDR:8445?encryption=none&security=reality&sni=$SNI&fp=chrome&pbk=$REALITY_PUBLIC_KEY&sid=$SHORT_ID&type=xhttp&path=/s/ref=nav_logo&mode=stream-one#Admin_XHTTP"

# Вывод
echo ""
echo "========== VLESS+REALITY LINKS =========="
echo ""
echo "🔵 TCP (port 8443) - for existing clients:"
echo "$TCP_LINK"
echo ""
echo "🟢 XHTTP (port 8445) - for modern clients:"
echo "$XHTTP_LINK"
echo ""
echo "=========================================="
echo ""

# Сохраняем в файлы
echo "$TCP_LINK" > /tmp/vless-tcp-link.txt
echo "$XHTTP_LINK" > /tmp/vless-xhttp-link.txt
echo "✅ TCP link saved to /tmp/vless-tcp-link.txt"
echo "✅ XHTTP link saved to /tmp/vless-xhttp-link.txt"
