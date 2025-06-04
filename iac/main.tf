resource "docker_image" "alcatraz_app" {
  name         = var.app_image
  keep_locally = true
}

# Pull Caddy image for load balancer
resource "docker_image" "caddy_lb" {
  name         = "caddy:alpine"
  keep_locally = true
}

# Create a custom network for the application
resource "docker_network" "alcatraz_network" {
  name   = "alcatraz-network"
  driver = "bridge"
}

resource "local_file" "app_config" {
  count    = var.app_node_count
  filename = "${path.module}/configs/server-${count.index + 1}-config.yaml"
  content = templatefile("${path.module}/templates/app-config.yaml.tpl", {
    listen_address = "0.0.0.0"
    port           = var.app_port
    cert_path      = "/etc/ssl/certs/server.crt"
    key_path       = "/etc/ssl/private/server.key"
    ca_path        = "/etc/ssl/ca/ca.crt"
  })
}

# TLS certificate files for application nodes
resource "local_file" "app_cert_file" {
  count    = var.app_node_count
  filename = "${path.module}/configs/server-${count.index + 1}-cert.crt"
  content  = tls_locally_signed_cert.app_cert[count.index].cert_pem
}

resource "local_file" "app_key_file" {
  count    = var.app_node_count
  filename = "${path.module}/configs/server-${count.index + 1}-key.key"
  content  = tls_private_key.app_key[count.index].private_key_pem
}

resource "local_file" "app_ca_file" {
  count    = var.app_node_count
  filename = "${path.module}/configs/server-${count.index + 1}-ca.crt"
  content  = tls_self_signed_cert.ca_cert.cert_pem
}

# Application containers with mTLS
resource "docker_container" "alcatraz_app" {
  count    = var.app_node_count
  image    = docker_image.alcatraz_app.image_id
  name     = "alcatraz-server-${count.index + 1}"
  hostname = "alcatraz-server-${count.index + 1}"

  networks_advanced {
    name = docker_network.alcatraz_network.name
  }

  ports {
    internal = var.app_port
  }

  # Mount configuration
  upload {
    file    = "/app/config.yaml"
    content = local_file.app_config[count.index].content
  }

  # Mount TLS certificates for mTLS
  upload {
    file    = "/etc/ssl/certs/server.crt"
    content = local_file.app_cert_file[count.index].content
  }

  upload {
    file    = "/etc/ssl/private/server.key"
    content = local_file.app_key_file[count.index].content
  }

  upload {
    file    = "/etc/ssl/ca/ca.crt"
    content = local_file.app_ca_file[count.index].content
  }

  # Pass config file as command line argument
  command = ["--config", "/app/config.yaml"]

  restart = "unless-stopped"

  # Disable health check since mTLS requires certificates that
  # aren't available to health check commands
  # Caddy will handle health checking of the backends
  healthcheck {
    test = ["NONE"]
  }

  labels {
    label = "service"
    value = "alcatraz-app"
  }

  labels {
    label = "node_id"
    value = tostring(count.index + 1)
  }
}
