locals {
  function_name = "${var.project_name}-lambda"
  bucket_name   = "${var.project_name}-artifacts"

  common_tags = merge(
    {
      Project     = var.project_name
      ManagedBy   = "Terraform"
      Environment = "production"
    },
    var.tags
  )
}

resource "aws_s3_bucket" "lambda_artifacts" {
  bucket = local.bucket_name

  tags = local.common_tags
}

resource "aws_s3_bucket_versioning" "lambda_artifacts" {
  bucket = aws_s3_bucket.lambda_artifacts.id

  versioning_configuration {
    status = "Enabled"
  }
}

resource "aws_s3_bucket_lifecycle_configuration" "lambda_artifacts" {
  bucket = aws_s3_bucket.lambda_artifacts.id

  rule {
    id     = "expire-old-lambda-artifacts"
    status = "Enabled"

    filter {}

    noncurrent_version_expiration {
      noncurrent_days = 30
    }
  }
}

resource "aws_s3_bucket_public_access_block" "lambda_artifacts" {
  bucket = aws_s3_bucket.lambda_artifacts.id

  block_public_acls       = true
  block_public_policy     = true
  ignore_public_acls      = true
  restrict_public_buckets = true
}

resource "aws_secretsmanager_secret" "discord_webhook" {
  name        = "${var.project_name}/discord-webhook-url"
  description = "Discord webhook URL for ${var.project_name}"

  tags = local.common_tags
}

resource "aws_iam_role" "lambda_execution" {
  name = "${var.project_name}-lambda-execution-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "lambda.amazonaws.com"
        }
      }
    ]
  })

  tags = local.common_tags
}

resource "aws_iam_role_policy_attachment" "lambda_basic_execution" {
  role       = aws_iam_role.lambda_execution.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}

resource "aws_iam_role_policy" "lambda_secrets_manager" {
  name = "${var.project_name}-secrets-manager-access"
  role = aws_iam_role.lambda_execution.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "secretsmanager:GetSecretValue"
        ]
        Resource = aws_secretsmanager_secret.discord_webhook.arn
      }
    ]
  })
}

resource "aws_cloudwatch_log_group" "lambda" {
  name              = "/aws/lambda/${local.function_name}"
  retention_in_days = var.log_retention_days

  tags = local.common_tags
}

resource "aws_s3_object" "lambda_package" {
  bucket = aws_s3_bucket.lambda_artifacts.id
  key    = "lambda.zip"
  source = "../lambda.zip"
  etag   = filemd5("../lambda.zip")

  tags = local.common_tags
}

resource "aws_lambda_function" "notification" {
  function_name = local.function_name
  role          = aws_iam_role.lambda_execution.arn
  handler       = "bootstrap"
  runtime       = "provided.al2023"
  architectures = ["arm64"]

  s3_bucket        = aws_s3_bucket.lambda_artifacts.id
  s3_key           = aws_s3_object.lambda_package.key
  source_code_hash = filebase64sha256("../lambda.zip")

  memory_size = var.lambda_memory_size
  timeout     = var.lambda_timeout

  environment {
    variables = {
      SECRET_ARN = aws_secretsmanager_secret.discord_webhook.arn
    }
  }

  depends_on = [
    aws_cloudwatch_log_group.lambda,
    aws_iam_role_policy_attachment.lambda_basic_execution,
    aws_iam_role_policy.lambda_secrets_manager
  ]

  tags = local.common_tags
}

resource "aws_iam_role" "scheduler_execution" {
  name = "${var.project_name}-scheduler-execution-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "scheduler.amazonaws.com"
        }
      }
    ]
  })

  tags = local.common_tags
}

resource "aws_iam_role_policy" "scheduler_lambda_invoke" {
  name = "${var.project_name}-scheduler-lambda-invoke"
  role = aws_iam_role.scheduler_execution.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = ["lambda:InvokeFunction"]
        Resource = aws_lambda_function.notification.arn
      }
    ]
  })
}

resource "aws_scheduler_schedule" "schedule" {
  name        = "${var.project_name}-schedule"
  description = "Trigger Lambda function daily at 6AM JST (21:00 UTC)"

  flexible_time_window {
    mode = "OFF"
  }

  schedule_expression          = var.schedule_expression
  schedule_expression_timezone = "Asia/Tokyo"

  target {
    arn      = aws_lambda_function.notification.arn
    role_arn = aws_iam_role.scheduler_execution.arn
  }
}
