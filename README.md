# gostfix

gostfix is simple go-based mail-manager for postfix with web interface

Supported features:

- ~Web admin interface~
- Web mail interface
- ~gRPC admin interface~
- ~POP3 inteface~
- ~IMAP interface~
- SASL authentication

# Prerequesties

gostfix only works on Linux-like operating systems

- go 1.13 or higher
- mongo 3.6 or higher
- protobuf 3.6.1 or higher
- nginx 1.18.0 or higher with websockets support

# Installation and setup

> TODO: Will be described later

# Nginx

```
    listen 443 ssl;
    server_name mail.example.com;

    # Add proxy micro-web services
    location / {
        proxy_pass http://localhost:65200;
    }

    # Add web sockets proxy
    location ~ ^/m/[\d]+/notifierSubscribe$ {
        proxy_pass http://localhost:65200;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "Upgrade";
        proxy_set_header Host $host;
    }


    # SSL configuration
    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/privkey.pem;
```
