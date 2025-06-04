# Global options
{
    admin localhost:2019
    auto_https off
}

# HTTP to HTTPS redirect
${domain_name}:80, localhost:80 {
    redir https://{host}{uri} permanent
}

# HTTPS with mTLS to backends
${domain_name}:443, localhost:443 {
    # TLS configuration
    tls /etc/ssl/certs/server.crt /etc/ssl/private/server.key

    # Health check endpoint
    handle /health {
        respond "healthy" 200
    }

    # Reverse proxy with mTLS to application nodes
    reverse_proxy {
        %{ for node in app_nodes ~}
        to https://${node.name}:${node.port}
        %{ endfor ~}
        
        # mTLS configuration for backend communication
        transport http {
            tls_client_auth /etc/ssl/certs/server.crt /etc/ssl/private/server.key
            tls_trusted_ca_certs /etc/ssl/ca/ca.crt
        }
        
        # Load balancing
        lb_policy least_conn
        
        # Health checks
        health_uri /api/ping
        health_interval 30s
        health_timeout 10s
    }

    # Logging
    log {
        output stdout
        format json
    }
} 