output "lambda_function_name" {
  description = "Name of the Lambda function"
  value       = aws_lambda_function.notification.function_name
}

output "lambda_function_arn" {
  description = "ARN of the Lambda function"
  value       = aws_lambda_function.notification.arn
}

output "eventbridge_rule_name" {
  description = "Name of the EventBridge rule"
  value       = aws_cloudwatch_event_rule.schedule.name
}

output "s3_bucket_name" {
  description = "Name of the S3 bucket for Lambda artifacts"
  value       = aws_s3_bucket.lambda_artifacts.id
}

output "cloudwatch_log_group" {
  description = "CloudWatch log group name"
  value       = aws_cloudwatch_log_group.lambda.name
}

output "discord_webhook_secret_arn" {
  description = "ARN of the Secrets Manager secret for Discord webhook URL"
  value       = aws_secretsmanager_secret.discord_webhook.arn
}
