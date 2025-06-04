

# Generate CA private key for mTLS
resource "tls_private_key" "ca_key" {
  algorithm = "RSA"
  rsa_bits  = 2048
}

# Generate CA certificate for mTLS
resource "tls_self_signed_cert" "ca_cert" {
  private_key_pem = tls_private_key.ca_key.private_key_pem

  subject {
    common_name  = "Alcatraz CA"
    organization = "Alcatraz Rest"
  }

  validity_period_hours = 8760 # 1 year

  is_ca_certificate = true

  allowed_uses = [
    "cert_signing",
    "key_encipherment",
    "digital_signature",
  ]
}

# Generate private keys for application nodes
resource "tls_private_key" "app_key" {
  count     = var.app_node_count
  algorithm = "RSA"
  rsa_bits  = 2048
}

# Generate certificate requests for application nodes
resource "tls_cert_request" "app_csr" {
  count           = var.app_node_count
  private_key_pem = tls_private_key.app_key[count.index].private_key_pem

  subject {
    common_name  = "alcatraz-server-${count.index + 1}"
    organization = "Alcatraz Rest"
  }

  dns_names = [
    "alcatraz-server-${count.index + 1}",
    "localhost",
    var.domain_name,
  ]
}

# Generate certificates for application nodes signed by CA
resource "tls_locally_signed_cert" "app_cert" {
  count              = var.app_node_count
  cert_request_pem   = tls_cert_request.app_csr[count.index].cert_request_pem
  ca_private_key_pem = tls_private_key.ca_key.private_key_pem
  ca_cert_pem        = tls_self_signed_cert.ca_cert.cert_pem

  validity_period_hours = 8760 # 1 year

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
    "client_auth",
  ]
}

# Generate private key for Caddy
resource "tls_private_key" "caddy_key" {
  algorithm = "RSA"
  rsa_bits  = 2048
}

# Generate certificate request for Caddy
resource "tls_cert_request" "caddy_csr" {
  private_key_pem = tls_private_key.caddy_key.private_key_pem

  subject {
    common_name  = var.domain_name
    organization = "Alcatraz Rest"
  }

  dns_names = [
    var.domain_name,
    "localhost",
    "alcatraz-load-balancer",
  ]
}

# Generate certificate for Caddy signed by CA
resource "tls_locally_signed_cert" "caddy_cert" {
  cert_request_pem   = tls_cert_request.caddy_csr.cert_request_pem
  ca_private_key_pem = tls_private_key.ca_key.private_key_pem
  ca_cert_pem        = tls_self_signed_cert.ca_cert.cert_pem

  validity_period_hours = 8760 # 1 year

  allowed_uses = [
    "key_encipherment",
    "digital_signature",
    "server_auth",
    "client_auth",
  ]
}
