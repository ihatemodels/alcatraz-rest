variable "app_node_count" {
  description = "Number of application nodes to deploy"
  type        = number
  default     = 3
}

variable "app_image" {
  description = "Docker image for the application"
  type        = string
  default     = "ghcr.io/ihatemodels/alcatraz-rest:latest"
}

variable "app_port" {
  description = "Port on which the application runs"
  type        = number
  default     = 9080
}

variable "lb_port" {
  description = "Load balancer port (Caddy will handle both HTTP and HTTPS)"
  type        = number
  default     = 80
}

variable "lb_https_port" {
  description = "Load balancer HTTPS port"
  type        = number
  default     = 443
}

variable "domain_name" {
  description = "Domain name for TLS certificate"
  type        = string
  default     = "alcatraz.rest"
}
