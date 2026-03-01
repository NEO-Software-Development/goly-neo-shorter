output "goly_service_ip" {
  description = "The public IP address of the Goly service."
  value       = kubernetes_service.goly_service.status[0].load_balancer.ingress[0].ip
}
