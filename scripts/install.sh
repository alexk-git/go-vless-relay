#!/bin/bash

set -e

echo "=== Installing VLESS+REALITY Server ==="

# Build
make build

# Install
sudo make install

# Generate REALITY keys if needed
if [ ! -f /etc/vless-server/.private_key ]; then
    echo "Generating REALITY keys..."
    openssl rand -base64 32 > /etc/vless-server/.private_key
fi

# Create first user if no users exist
if [ ! -f /var/lib/vless-server/clients.json ]; then
    echo "Creating first user..."
    /usr/local/bin/vless-server -add-user "Admin" -email "admin@localhost"
fi

# Enable and start service
sudo systemctl enable vless-server
sudo systemctl start vless-server

echo "=== Installation complete ==="
echo "Check status: sudo systemctl status vless-server"
echo "View logs: sudo journalctl -u vless-server -f"
