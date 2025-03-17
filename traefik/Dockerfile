FROM traefik:v3.3.1

# needed for health-check
RUN apk add curl

RUN [ -f /etc/traefik/acme/acme.json ] && chmod 600 /etc/traefik/acme/acme.json || echo "acme.json not found, skipping"