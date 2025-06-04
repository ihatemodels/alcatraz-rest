# Generate Caddyfile configuration
resource "local_file" "caddyfile" {
  filename = "${path.module}/configs/Caddyfile"
  content = templatefile("${path.module}/templates/Caddyfile.tpl", {
    domain_name = var.domain_name
    app_nodes = [for i in range(var.app_node_count) : {
      name = "alcatraz-server-${i + 1}"
      port = var.app_port
    }]
  })
}

# TLS certificate files for Caddy
resource "local_file" "caddy_cert_file" {
  filename = "${path.module}/configs/caddy-cert.crt"
  content  = tls_locally_signed_cert.caddy_cert.cert_pem
}

resource "local_file" "caddy_key_file" {
  filename = "${path.module}/configs/caddy-key.key"
  content  = tls_private_key.caddy_key.private_key_pem
}

resource "local_file" "caddy_ca_file" {
  filename = "${path.module}/configs/caddy-ca.crt"
  content  = tls_self_signed_cert.ca_cert.cert_pem
}

# Load balancer container using Caddy
resource "docker_container" "caddy_lb" {
  image = docker_image.caddy_lb.image_id
  name  = "alcatraz-load-balancer"

  networks_advanced {
    name = docker_network.alcatraz_network.name
  }

  ports {
    internal = var.lb_port
    external = var.lb_port
  }

  ports {
    internal = var.lb_https_port
    external = var.lb_https_port
  }

  # Mount Caddyfile configuration
  upload {
    file    = "/etc/caddy/Caddyfile"
    content = local_file.caddyfile.content
  }

  # Mount TLS certificates for mTLS
  upload {
    file    = "/etc/ssl/certs/server.crt"
    content = local_file.caddy_cert_file.content
  }

  upload {
    file    = "/etc/ssl/private/server.key"
    content = local_file.caddy_key_file.content
  }

  upload {
    file    = "/etc/ssl/ca/ca.crt"
    content = local_file.caddy_ca_file.content
  }

  restart = "unless-stopped"

  healthcheck {
    test = [
      "CMD",
      "wget",
      "--quiet",
      "--tries=1",
      "--spider",
      "http://localhost:2019/metrics"
    ]
    interval = "30s"
    timeout  = "10s"
    retries  = 3
  }

  labels {
    label = "service"
    value = "alcatraz-load-balancer"
  }

  depends_on = [docker_container.alcatraz_app]
}
