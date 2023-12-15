output "target_group_arn" {
  value       = aws_lb_target_group.front_end.arn
  description = "ARN of the load balancer target group."
}
