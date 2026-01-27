<!-- BEGIN_TF_DOCS -->
# Terraform Configuration

## Usage

This Terraform configuration manages AWS infrastructure for the shin-yokohama-event-notifier project.



## Requirements

| Name | Version |
|------|---------|
| <a name="requirement_terraform"></a> [terraform](#requirement\_terraform) | >= 1.0 |
| <a name="requirement_aws"></a> [aws](#requirement\_aws) | >= 5.0, < 6.0 |
| <a name="requirement_grafana"></a> [grafana](#requirement\_grafana) | >= 3.0 |

## Providers

| Name | Version |
|------|---------|
| <a name="provider_aws"></a> [aws](#provider\_aws) | 5.100.0 |
| <a name="provider_grafana"></a> [grafana](#provider\_grafana) | 4.25.0 |

## Inputs

| Name | Description | Type | Default | Required |
|------|-------------|------|---------|:--------:|
| <a name="input_aws_region"></a> [aws\_region](#input\_aws\_region) | AWS region for resource deployment | `string` | `"ap-northeast-1"` | no |
| <a name="input_grafana_auth"></a> [grafana\_auth](#input\_grafana\_auth) | Grafana Cloud Service Account Token | `string` | n/a | yes |
| <a name="input_grafana_url"></a> [grafana\_url](#input\_grafana\_url) | Grafana Cloud stack URL (e.g., https://your-stack.grafana.net) | `string` | n/a | yes |
| <a name="input_lambda_memory_size"></a> [lambda\_memory\_size](#input\_lambda\_memory\_size) | Memory size for Lambda function in MB | `number` | `128` | no |
| <a name="input_lambda_timeout"></a> [lambda\_timeout](#input\_lambda\_timeout) | Timeout for Lambda function in seconds | `number` | `30` | no |
| <a name="input_log_retention_days"></a> [log\_retention\_days](#input\_log\_retention\_days) | CloudWatch Logs retention period in days | `number` | `7` | no |
| <a name="input_project_name"></a> [project\_name](#input\_project\_name) | Project name used for resource naming | `string` | `"shin-yokohama-event-notifier"` | no |
| <a name="input_schedule_expression"></a> [schedule\_expression](#input\_schedule\_expression) | EventBridge cron schedule expression (UTC timezone) | `string` | `"cron(0 0 1 1 ? 2099)"` | no |
| <a name="input_tags"></a> [tags](#input\_tags) | Additional tags to apply to resources | `map(string)` | `{}` | no |

## Outputs

| Name | Description |
|------|-------------|
| <a name="output_cloudwatch_log_group"></a> [cloudwatch\_log\_group](#output\_cloudwatch\_log\_group) | CloudWatch log group name |
| <a name="output_discord_webhook_secret_arn"></a> [discord\_webhook\_secret\_arn](#output\_discord\_webhook\_secret\_arn) | ARN of the Secrets Manager secret for Discord webhook URL |
| <a name="output_eventbridge_rule_name"></a> [eventbridge\_rule\_name](#output\_eventbridge\_rule\_name) | Name of the EventBridge rule |
| <a name="output_lambda_function_arn"></a> [lambda\_function\_arn](#output\_lambda\_function\_arn) | ARN of the Lambda function |
| <a name="output_lambda_function_name"></a> [lambda\_function\_name](#output\_lambda\_function\_name) | Name of the Lambda function |
| <a name="output_s3_bucket_name"></a> [s3\_bucket\_name](#output\_s3\_bucket\_name) | Name of the S3 bucket for Lambda artifacts |
<!-- END_TF_DOCS -->