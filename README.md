# Snarky
### Zero-Knowledge, Asynchronous Dead Drop

![Snarky Logo](https://sapadian.com/assets/snarky_logo.png)

***Snarky*** is a self-hosted, Zero-Knowledge file transfer tool designed for DevOps teams and Sysadmins. It solves the problem: "I need to send you a sensitive file (archives, images, PDFs, keys), but you aren't online, and I don't want it stored permanently on a server."

---

## 🚀 Features

* **Zero-Knowledge:** Encryption happens on the client before upload. The server never sees the encryption key, the filename, or the file type.
* **Burn-After-Reading:** The moment a file is downloaded, it is scrubbed from the server instantly.
* **Encrypted-at-Rest:** Data is stored on the server's disk as opaque, encrypted blobs. Even with root access to the server, the files are unreadable.
* **Ephemeral:** Files not picked up within the configured retention period (default: 24h) are automatically incinerated.
* **Binary Support:** Now supports generic file transfers up to 10MB (configurable), including images, PDFs, and ZIP archives.

---

## 📦 Client Installation

You can download pre-compiled binaries from the [Releases Page](https://sapadian.com/snarky/clients) or build it yourself.

### Build from Source

```bash
git clone https://github.com/sapadianllc/snarky.git
cd snarky
go build -o snarky main.go
```

---

## 🚀 Quick Test (Public Server)

Try Snarky immediately using the free public instance hosted by Sapadian:

```bash
# Send a file
./snarky send -file secret.txt -host https://snarkypub.sapadian.com

# Get a file
./snarky get -id <ID> -key <KEY> -host https://snarkypub.sapadian.com
```

---

# 🖥️ Server Deployment (FreeBSD)

Follow these steps to host your own private instance of Snarky.

## 1. Prerequisites

- A FreeBSD Server (VPS or Physical)  
- Go 1.21+ installed (`pkg install go`)  
- Root access  

## 2. Build the Daemon

Compile the server binary specifically for FreeBSD.

```bash
# If building on the server itself:
go build -o snarky-bsd main.go

# Move to bin directory
sudo mv snarky-bsd /usr/local/bin/snarky
sudo chmod +x /usr/local/bin/snarky

#create folders for file storage
sudo mkdir -p /var/db/snarky
sudo chown -R nobody:nobody /var/db/snarky
sudo chmod 750 /var/db/snarky
```

## 3. Create the configuration file to control storage and limits.

Create the file: /usr/local/etc/snarky.json

```json
{
  "port": "8080",
  "storage_path": "/var/db/snarky",
  "max_file_size": 10485760,
  "retention": "24h"
}
```
Note: 10485760 bytes is 10MB.

## 4. Setup the Service (RC.D)

Create the service script to manage Snarky as a background daemon.

Create the file: `/usr/local/etc/rc.d/snarky`

```bash
#!/bin/sh

# PROVIDE: snarky
# REQUIRE: NETWORKING
# KEYWORD: shutdown

. /etc/rc.subr

name="snarky"
rcvar="snarky_enable"

load_rc_config $name

: ${snarky_enable:="NO"}
: ${snarky_user:="nobody"}
: ${snarky_group:="nobody"}
: ${snarky_config:="/usr/local/etc/snarky.json"}

pidfile="/var/run/snarky/${name}.pid"
procname="/usr/local/bin/snarky"
command="/usr/sbin/daemon"

start_precmd="snarky_precmd"

snarky_precmd()
{
    # Ensure PID directory exists
    if [ ! -d "/var/run/snarky" ]; then
        install -d -m 0755 -o ${snarky_user} -g ${snarky_group} /var/run/snarky
    fi
    # Ensure Storage directory exists (as defined in json)
    if [ ! -d "/var/db/snarky" ]; then
        install -d -m 0755 -o ${snarky_user} -g ${snarky_group} /var/db/snarky
    fi
}

command_args="-P ${pidfile} -r -u ${snarky_user} -f ${procname} server -config ${snarky_config}"

run_rc_command "$1"
```

Enable and start the service:

```bash
sudo chmod +x /usr/local/etc/rc.d/snarky
sudo sysrc snarky_enable="YES"
sudo service snarky start
```

---

# 🔒 Nginx Reverse Proxy & SSL (HTTPS)

It is highly recommended to put Nginx in front of Snarky to handle SSL/TLS.

## 1. Install Nginx & Certbot

```bash
sudo pkg update
sudo pkg install nginx py311-certbot py311-certbot-nginx
sudo sysrc nginx_enable="YES"
```

## 2. Configure Nginx

Edit /usr/local/etc/nginx/nginx.conf. Important: Ensure client_max_body_size matches or exceeds the limit set in your snarky.json.

```nginx
server {
    listen 80;
    server_name snarky.yourdomain.com;

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        client_max_body_size 100M;
    }
}
```

## 3. Enable SSL (Auto-Cert)

```bash
sudo service nginx start
sudo certbot --nginx -d snarky.yourdomain.com
```

Choose the option to redirect all HTTP traffic to HTTPS.

## 4. Setup Auto-Renewal

```bash
sudo crontab -e
```

Add:

```bash
0 0,12 * * * /usr/local/bin/certbot renew --quiet --deploy-hook "service nginx reload"
```

---

# 💻 Usage Guide

## Sending a Secret

```bash
snarky send -file ./database_creds.env -host https://snarky.yourdomain.com
```

Example output:

```
Reading database_creds.env...
Encrypting... Done.
Uploading to Dead Drop...
Progress: [====================] 100.00%

[SECURE DROP CREATED]
ID:  a1b2c3d4-5678...
KEY: XyZ123_SecretKey...

To retrieve:
snarky get -id a1b2c3d4... -key XyZ123... -host https://snarky.yourdomain.com
```

## Receiving a Secret

```bash
snarky get -id <ID> -key <KEY> -host https://snarky.yourdomain.com
```
Output:

```
Connecting to Dead Drop...
Decrypting... Done.
[SUCCESS] Saved as: downloaded_database_creds.env
```

---

# 📄 License

Snarky is licensed under the GNU AGPLv3.

You are free to use and modify this software.

If you host Snarky as a public service, you must make the source code of your version available to users.




