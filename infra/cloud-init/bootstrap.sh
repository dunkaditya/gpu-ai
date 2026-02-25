#!/bin/bash
# GPU.ai cloud-init bootstrap
# Injected into upstream instance at provisioning time

set -euo pipefail

# Variables injected by provisioning service
INSTANCE_ID="{{instance_id}}"
PROXY_ENDPOINT="{{proxy_public_ip}}"
PROXY_PUBLIC_KEY="{{proxy_wireguard_public_key}}"
INSTANCE_PRIVATE_KEY="{{instance_wireguard_private_key}}"
INSTANCE_ADDRESS="{{instance_wireguard_address}}"  # e.g., 10.0.0.x/24
SSH_AUTHORIZED_KEYS="{{ssh_authorized_keys}}"
DOCKER_IMAGE="{{docker_image}}"  # optional

# 1. Install WireGuard
apt-get update -qq && apt-get install -y -qq wireguard

# 2. Configure WireGuard tunnel
cat > /etc/wireguard/wg0.conf << EOF
[Interface]
PrivateKey = ${INSTANCE_PRIVATE_KEY}
Address = ${INSTANCE_ADDRESS}

[Peer]
PublicKey = ${PROXY_PUBLIC_KEY}
Endpoint = ${PROXY_ENDPOINT}:51820
AllowedIPs = 10.0.0.0/24
PersistentKeepalive = 25
EOF

chmod 600 /etc/wireguard/wg0.conf
systemctl enable wg-quick@wg0
systemctl start wg-quick@wg0

# 3. Set hostname
hostnamectl set-hostname "gpu-${INSTANCE_ID}.gpu.ai"

# 4. Set MOTD
cat > /etc/motd << 'MOTDEOF'

  ██████╗ ██████╗ ██╗   ██╗   █████╗ ██╗
 ██╔════╝ ██╔══██╗██║   ██║  ██╔══██╗██║
 ██║  ███╗██████╔╝██║   ██║  ███████║██║
 ██║   ██║██╔═══╝ ██║   ██║  ██╔══██║██║
 ╚██████╔╝██║     ╚██████╔╝  ██║  ██║██║
  ╚═════╝ ╚═╝      ╚═════╝   ╚═╝  ╚═╝╚═╝

 Welcome to GPU.ai
 Instance: gpu-${INSTANCE_ID}
 Support: support@gpu.ai

MOTDEOF

# 5. Configure SSH keys
mkdir -p /root/.ssh
echo "${SSH_AUTHORIZED_KEYS}" > /root/.ssh/authorized_keys
chmod 700 /root/.ssh
chmod 600 /root/.ssh/authorized_keys

# 6. Firewall — only allow traffic via WireGuard tunnel
iptables -A INPUT -i wg0 -j ACCEPT
iptables -A INPUT -p udp --dport 51820 -j ACCEPT  # WireGuard handshake
iptables -A INPUT -i lo -j ACCEPT
iptables -A INPUT -m state --state ESTABLISHED,RELATED -j ACCEPT
iptables -A INPUT -j DROP  # block everything else (including direct SSH)

# 7. Optional: pull and run Docker image
if [ -n "${DOCKER_IMAGE}" ]; then
    docker pull "${DOCKER_IMAGE}"
    docker run -d --gpus all --name workspace "${DOCKER_IMAGE}"
fi

# 8. Signal ready (call back to GPU.ai API)
curl -s -X POST "https://api.gpu.ai/internal/instances/${INSTANCE_ID}/ready" \
    -H "Authorization: Bearer {{internal_api_token}}"
