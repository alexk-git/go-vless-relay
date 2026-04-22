#!/bin/bash

set -e

echo "=== Uninstalling VLESS+REALITY Server ==="

sudo systemctl stop vless-server || true
sudo make uninstall

echo "=== Uninstall complete ==="
