# Nginx Setup

This guide sets up Nginx as a reverse proxy in front of Kasku so you can access the app via `http://202.10.42.99` (port 80) instead of `http://202.10.42.99:8080`.

## Prerequisites

- Ubuntu server with Kasku deployed (`make deploy`)
- Nginx installed:
  ```bash
  sudo apt-get update && sudo apt-get install -y nginx
  ```

## Install the config

```bash
# Copy the config file
sudo cp deploy/nginx.conf /etc/nginx/sites-available/kasku

# Enable it
sudo ln -sf /etc/nginx/sites-available/kasku /etc/nginx/sites-enabled/kasku

# (Optional) Remove the default Nginx welcome page
sudo rm -f /etc/nginx/sites-enabled/default

# Test and reload
sudo nginx -t && sudo systemctl reload nginx
```

## Update APP_URL

After setting up Nginx, update `/opt/kasku/.env` so invite links use the public URL:

```bash
# Edit the env file
sudo nano /opt/kasku/.env
```

Change:
```
APP_URL=http://localhost:8080
```
To:
```
APP_URL=http://202.10.42.99
```

Then restart Kasku:
```bash
sudo systemctl restart kasku
```

## Verify

```bash
# Check Nginx status
sudo systemctl status nginx

# Test the proxy
curl -I http://202.10.42.99
```

You should see a `200 OK` or a `302` redirect to `/login`.

## Firewall

If you use `ufw`, allow HTTP traffic:

```bash
sudo ufw allow 'Nginx HTTP'
sudo ufw status
```

You can optionally block direct access to port 8080 from outside, since Nginx proxies on port 80:

```bash
sudo ufw deny 8080
```

## Troubleshooting

| Problem | Check |
|---------|-------|
| 502 Bad Gateway | Is Kasku running? `sudo systemctl status kasku` |
| Connection refused | Is Nginx running? `sudo systemctl status nginx` |
| Config syntax error | Run `sudo nginx -t` to see details |
| Logs | `sudo tail -f /var/log/nginx/error.log` |

## Next Steps (Phase 2)

When you add a domain name, you can:

1. Update `server_name` in the Nginx config to your domain
2. Install a free TLS certificate with Certbot:
   ```bash
   sudo apt-get install -y certbot python3-certbot-nginx
   sudo certbot --nginx -d yourdomain.com
   ```
3. Update `APP_URL` in `/opt/kasku/.env` to `https://yourdomain.com`
