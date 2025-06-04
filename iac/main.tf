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
