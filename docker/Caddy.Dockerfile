FROM caddy:2.7.6-builder AS builder

RUN xcaddy build \
    --with github.com/caddyserver/certmagic@v0.20.0 \
    --with github.com/caddy-dns/cloudflare


FROM caddy:2.7.6-alpine AS runner

COPY --from=builder /usr/bin/caddy /usr/bin/caddy