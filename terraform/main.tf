terraform {
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 5.0"
    }
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
  zone    = var.zone
}

# 1. Regra de Firewall: Abre a porta 8080 (API) e 22 (SSH)
resource "google_compute_firewall" "allow_api" {
  name    = "allow-auth-api"
  network = "default"

  allow {
    protocol = "tcp"
    ports    = ["8080", "22"]
  }

  source_ranges = ["0.0.0.0/0"] # Permite tráfego de qualquer IP
  target_tags   = ["auth-api"]
}

# 2. A Máquina Virtual (VM)
resource "google_compute_instance" "auth_server" {
  name         = "api-auth-server"
  machine_type = "e2-micro" # 🚨 Instância do Free Tier
  zone         = var.zone

  boot_disk {
    initialize_params {
      image = "ubuntu-os-cloud/ubuntu-2204-lts"
      size  = 30 # 🚨 Limite do Free Tier (30GB Standard)
      type  = "pd-standard"
    }
  }

  network_interface {
    network = "default"
    access_config {
      # Deixar este bloco vazio cria um IP Público dinâmico automaticamente
    }
  }

  # Essa tag conecta a VM à regra de firewall criada acima
  tags = ["auth-api"]
}

# 3. Output: Mostra o IP no terminal após criar a máquina
output "public_ip" {
  value       = google_compute_instance.auth_server.network_interface[0].access_config[0].nat_ip
  description = "O IP público do seu servidor"
}