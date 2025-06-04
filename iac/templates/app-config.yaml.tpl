server:
  listen_address: "${listen_address}"
  port: ${port}
  tls:
    enabled: true
    cert_file: "${cert_path}"
    key_file: "${key_path}"
    client_ca_file: "${ca_path}"
    require_client_cert: true
log:
  level: "info"
  type: "json" 