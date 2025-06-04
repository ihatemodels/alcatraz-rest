output "load_balancer_http_url" {
  description = "HTTP URL for the load balancer (redirects to HTTPS)"
  value       = "http://localhost:${var.lb_port}"
}

output "load_balancer_https_url" {
  description = "HTTPS URL for the load balancer with mTLS"
  value       = "https://localhost:${var.lb_https_port}"
}

output "app_node_info" {
  description = "Application node information (accessible only through load balancer)"
  value = [
    for i in range(var.app_node_count) : {
      node_id        = i + 1
      container_name = "alcatraz-server-${i + 1}"
      internal_url   = "https://alcatraz-server-${i + 1}:${var.app_port}"
      note           = "Accessible only through load balancer or internal Docker network"
    }
  ]
}

output "network_name" {
  description = "Docker network name"
  value       = docker_network.alcatraz_network.name
}

output "app_node_count" {
  description = "Number of deployed application nodes"
  value       = var.app_node_count
}

output "mtls_enabled" {
  description = "mTLS is always enabled in this configuration"
  value       = true
}

output "container_names" {
  description = "Names of all deployed containers"
  value = {
    load_balancer = docker_container.caddy_lb.name
    app_nodes     = [for container in docker_container.alcatraz_app : container.name]
  }
}

output "health_check_urls" {
  description = "Health check endpoints"
  value = {
    load_balancer_health = "http://localhost:${var.lb_port}/health"
    load_balancer_https  = "https://localhost:${var.lb_https_port}/api/ping"
    caddy_admin          = "http://localhost:2019/metrics"
    note                 = "App nodes are health-checked internally by Caddy"
  }
}
