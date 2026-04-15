variable "aws_region" {
  description = "AWS region for resource deployment"
  type        = string
  default     = "ap-northeast-1"
}

variable "project_name" {
  description = "Project name used for resource naming"
  type        = string
  default     = "shin-yokohama-event-notifier"
}

variable "lambda_memory_size" {
  description = "Memory size for Lambda function in MB"
  type        = number
  default     = 128
}

variable "lambda_timeout" {
  description = "Timeout for Lambda function in seconds"
  type        = number
  default     = 30
}

variable "lambda_weekly_timeout" {
  description = "Timeout for weekly Lambda function in seconds"
  type        = number
  default     = 120
}

variable "schedule_expression" {
  description = "Amazon EventBridge Scheduler cron expression for triggering the notification workflow (Asia/Tokyo timezone)"
  type        = string
  default     = "cron(0 6 * * ? *)" # Daily at 6AM JST
}

variable "log_retention_days" {
  description = "CloudWatch Logs retention period in days"
  type        = number
  default     = 7
}

variable "tags" {
  description = "Additional tags to apply to resources"
  type        = map(string)
  default     = {}
}

variable "grafana_url" {
  description = "Grafana Cloud stack URL (e.g., https://your-stack.grafana.net)"
  type        = string
}

variable "grafana_auth" {
  description = "Grafana Cloud Service Account Token"
  type        = string
  sensitive   = true
}
