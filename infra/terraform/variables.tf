variable "app_image" {
  description = "The Docker image for the Goly application."
  type        = string
  default     = "goly-app:latest" # This is a placeholder. You will need to build and push your own image.
}

variable "replicas" {
  description = "The number of replicas for the Goly application."
  type        = number
  default     = 1
}

variable "db_host" {
  description = "The hostname of the PostgreSQL database."
  type        = string
}

variable "db_user" {
  description = "The username for the PostgreSQL database."
  type        = string
}

variable "db_password" {
  description = "The password for the PostgreSQL database."
  type        = string
  sensitive   = true
}

variable "db_name" {
  description = "The name of the PostgreSQL database."
  type        = string
}
