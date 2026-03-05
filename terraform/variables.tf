
variable "project_id" {
  description = "O ID do seu projeto no Google Cloud"
  type        = string
}

variable "region" {
  description = "Região elegível ao Free Tier"
  default     = "us-central1" 
}

variable "zone" {
  description = "Zona elegível ao Free Tier"
  default     = "us-central1-a"
}