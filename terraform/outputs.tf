output "lambda_daily_function_name" {
  description = "Name of the daily Lambda function"
  value       = aws_lambda_function.notification_daily.function_name
}

output "lambda_daily_function_arn" {
  description = "ARN of the daily Lambda function"
  value       = aws_lambda_function.notification_daily.arn
}

output "lambda_weekly_function_name" {
  description = "Name of the weekly Lambda function"
  value       = aws_lambda_function.notification_weekly.function_name
}

output "lambda_weekly_function_arn" {
  description = "ARN of the weekly Lambda function"
  value       = aws_lambda_function.notification_weekly.arn
}

output "eventbridge_schedule_daily_name" {
  description = "Name of the daily EventBridge Scheduler schedule"
  value       = aws_scheduler_schedule.schedule_daily.name
}

output "eventbridge_schedule_weekly_name" {
  description = "Name of the weekly EventBridge Scheduler schedule"
  value       = aws_scheduler_schedule.schedule_weekly.name
}

output "s3_bucket_name" {
  description = "Name of the S3 bucket for Lambda artifacts"
  value       = aws_s3_bucket.lambda_artifacts.id
}

output "cloudwatch_log_group_daily" {
  description = "CloudWatch log group name for the daily Lambda"
  value       = aws_cloudwatch_log_group.lambda_daily.name
}

output "discord_webhook_secret_arn" {
  description = "ARN of the Secrets Manager secret for Discord webhook URL"
  value       = aws_secretsmanager_secret.discord_webhook.arn
}

output "grafana_dashboard_url" {
  description = "URL of the Grafana Lambda monitoring dashboard"
  value       = grafana_dashboard.lambda.url
}
