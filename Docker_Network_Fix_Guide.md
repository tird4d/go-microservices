# Docker Network Fix Guide - ERR_NETWORK_CHANGED

## Problem
When Docker starts, you get this error:
```
ERR_NETWORK_CHANGED
Sorry, your request failed. Please check your firewall rules and network connection then try again.
```

## Root Cause
Docker conflicts with system IPv6 and modifies the routing table, causing network disconnections.

---

## Quick Fix (Temporary)

### 1. Stop Docker
```bash
sudo systemctl stop docker docker.socket
```

### 2. Disable IPv6
```bash
sudo sysctl -w net.ipv6.conf.all.disable_ipv6=1
sudo sysctl -w net.ipv6.conf.default.disable_ipv6=1
```

### 3. Start Docker
```bash
sudo systemctl start docker
```

---

## Permanent Fix

### 1. Permanently Disable IPv6
```bash
echo -e "net.ipv6.conf.all.disable_ipv6=1\nnet.ipv6.conf.default.disable_ipv6=1\nnet.ipv6.conf.lo.disable_ipv6=1" | sudo tee /etc/sysctl.d/99-disable-ipv6.conf
sudo sysctl --system
```

### 2. Configure Docker Daemon
Edit `/etc/docker/daemon.json`:
```json
{
  "bip": "10.200.0.1/24",
  "default-address-pools": [
    {
      "base": "10.201.0.0/16",
      "size": 24
    }
  ],
  "iptables": true
}
```

### 3. Restart Docker
```bash
sudo systemctl restart docker
```

---

## Useful Troubleshooting Commands

### Check Docker Networks
```bash
docker network ls
```

### Check Routing Table
```bash
ip route show
```

### Check Network Bridges
```bash
ip link show type bridge
```

### Prune Unused Docker Networks
```bash
docker network prune -f
```

### Manually Delete Old Bridges
```bash
# First stop Docker
sudo systemctl stop docker docker.socket

# Then delete bridges
sudo ip link delete br-XXXXXX
```

### Check IPv6 Status
```bash
cat /proc/sys/net/ipv6/conf/all/disable_ipv6
# 1 = disabled (correct)
# 0 = enabled
```

---

## Important Notes

1. **Address Range:** Use `10.x.x.x` as it has less conflict with common networks (`172.x.x.x` and `192.168.x.x`).

2. **iptables:** Must be `true`, otherwise containers won't have internet access.

3. **After Reboot:** If you applied the permanent fix, settings will persist after reboot.

---

## History
- **Date:** December 11, 2025
- **Issue:** Docker IPv6 conflict causing ERR_NETWORK_CHANGED
- **Solution:** Disable IPv6 + Configure Docker address range
