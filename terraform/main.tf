locals {
  function_name_daily  = "${var.project_name}-lambda-daily"
  function_name_weekly = "${var.project_name}-lambda-weekly"
  state_machine_name   = "${var.project_name}-notification"
  bucket_name          = "${var.project_name}-artifacts"

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

resource "aws_cloudwatch_log_group" "lambda_daily" {
  name              = "/aws/lambda/${local.function_name_daily}"
  retention_in_days = var.log_retention_days

  tags = local.common_tags
}

resource "aws_cloudwatch_log_group" "lambda_weekly" {
  name              = "/aws/lambda/${local.function_name_weekly}"
  retention_in_days = var.log_retention_days

  tags = local.common_tags
}

resource "aws_s3_object" "lambda_daily_package" {
  bucket = aws_s3_bucket.lambda_artifacts.id
  key    = "lambda-daily.zip"
  source = "../lambda-daily.zip"
  etag   = filemd5("../lambda-daily.zip")

  tags = local.common_tags
}

resource "aws_s3_object" "lambda_weekly_package" {
  bucket = aws_s3_bucket.lambda_artifacts.id
  key    = "lambda-weekly.zip"
  source = "../lambda-weekly.zip"
  etag   = filemd5("../lambda-weekly.zip")

  tags = local.common_tags
}

resource "aws_lambda_function" "notification_daily" {
  function_name = local.function_name_daily
  role          = aws_iam_role.lambda_execution.arn
  handler       = "bootstrap"
  runtime       = "provided.al2023"
  architectures = ["arm64"]

  s3_bucket        = aws_s3_bucket.lambda_artifacts.id
  s3_key           = aws_s3_object.lambda_daily_package.key
  source_code_hash = filebase64sha256("../lambda-daily.zip")

  memory_size = var.lambda_memory_size
  timeout     = var.lambda_timeout

  environment {
    variables = {
      SECRET_ARN = aws_secretsmanager_secret.discord_webhook.arn
    }
  }

  depends_on = [
    aws_cloudwatch_log_group.lambda_daily,
    aws_iam_role_policy_attachment.lambda_basic_execution,
    aws_iam_role_policy.lambda_secrets_manager
  ]

  tags = local.common_tags
}

resource "aws_lambda_function" "notification_weekly" {
  function_name = local.function_name_weekly
  role          = aws_iam_role.lambda_execution.arn
  handler       = "bootstrap"
  runtime       = "provided.al2023"
  architectures = ["arm64"]

  s3_bucket        = aws_s3_bucket.lambda_artifacts.id
  s3_key           = aws_s3_object.lambda_weekly_package.key
  source_code_hash = filebase64sha256("../lambda-weekly.zip")

  memory_size = var.lambda_memory_size
  timeout     = var.lambda_weekly_timeout

  environment {
    variables = {
      SECRET_ARN = aws_secretsmanager_secret.discord_webhook.arn
    }
  }

  depends_on = [
    aws_cloudwatch_log_group.lambda_weekly,
    aws_iam_role_policy_attachment.lambda_basic_execution,
    aws_iam_role_policy.lambda_secrets_manager
  ]

  tags = local.common_tags
}

# -----------------------------------------------------------------------------
# Step Functions
# -----------------------------------------------------------------------------

resource "aws_cloudwatch_log_group" "sfn" {
  name              = "/aws/states/${local.state_machine_name}"
  retention_in_days = var.log_retention_days

  tags = local.common_tags
}

resource "aws_iam_role" "sfn_execution" {
  name = "${var.project_name}-sfn-execution-role"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Principal = {
          Service = "states.amazonaws.com"
        }
      }
    ]
  })

  tags = local.common_tags
}

resource "aws_iam_role_policy" "sfn_lambda_invoke" {
  name = "${var.project_name}-sfn-lambda-invoke"
  role = aws_iam_role.sfn_execution.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = ["lambda:InvokeFunction"]
        Resource = [
          aws_lambda_function.notification_daily.arn,
          aws_lambda_function.notification_weekly.arn,
        ]
      }
    ]
  })
}

resource "aws_iam_role_policy" "sfn_logging" {
  name = "${var.project_name}-sfn-logging"
  role = aws_iam_role.sfn_execution.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Action = [
          "logs:CreateLogDelivery",
          "logs:CreateLogStream",
          "logs:GetLogDelivery",
          "logs:UpdateLogDelivery",
          "logs:DeleteLogDelivery",
          "logs:ListLogDeliveries",
          "logs:PutLogEvents",
          "logs:PutResourcePolicy",
          "logs:DescribeResourcePolicies",
          "logs:DescribeLogGroups"
        ]
        Resource = "*"
      }
    ]
  })
}

resource "aws_sfn_state_machine" "notification" {
  name     = local.state_machine_name
  role_arn = aws_iam_role.sfn_execution.arn

  definition = jsonencode({
    Comment       = "Shin-Yokohama event notification workflow. Runs weekly (Monday only) then daily."
    QueryLanguage = "JSONata"
    StartAt       = "CheckDayOfWeek"
    States = {
      CheckDayOfWeek = {
        Type = "Pass"
        Assign = {
          dayOfWeek = "{% ($floor(($toMillis($now()) + 32400000) / 86400000) + 4) % 7 %}"
        }
        Next = "IsMonday"
      }
      IsMonday = {
        Type = "Choice"
        Choices = [
          {
            Condition = "{% $dayOfWeek = 1 %}"
            Next      = "RunWeekly"
          }
        ]
        Default = "RunDaily"
      }
      RunWeekly = {
        Type     = "Task"
        Resource = "arn:aws:states:::lambda:invoke"
        Arguments = {
          FunctionName = aws_lambda_function.notification_weekly.arn
        }
        Catch = [
          {
            ErrorEquals = ["States.ALL"]
            Next        = "RunDaily"
          }
        ]
        Next = "RunDaily"
      }
      RunDaily = {
        Type     = "Task"
        Resource = "arn:aws:states:::lambda:invoke"
        Arguments = {
          FunctionName = aws_lambda_function.notification_daily.arn
        }
        End = true
      }
    }
  })

  logging_configuration {
    log_destination        = "${aws_cloudwatch_log_group.sfn.arn}:*"
    include_execution_data = false
    level                  = "ERROR"
  }

  tags = local.common_tags
}

# -----------------------------------------------------------------------------
# EventBridge Scheduler
# -----------------------------------------------------------------------------

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

resource "aws_iam_role_policy" "scheduler_sfn_start" {
  name = "${var.project_name}-scheduler-sfn-start"
  role = aws_iam_role.scheduler_execution.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect   = "Allow"
        Action   = ["states:StartExecution"]
        Resource = aws_sfn_state_machine.notification.arn
      }
    ]
  })
}

resource "aws_scheduler_schedule" "notification" {
  name        = "${var.project_name}-schedule"
  description = "Trigger notification Step Functions workflow at 6AM JST daily"

  flexible_time_window {
    mode = "OFF"
  }

  schedule_expression          = var.schedule_expression
  schedule_expression_timezone = "Asia/Tokyo"

  target {
    arn      = aws_sfn_state_machine.notification.arn
    role_arn = aws_iam_role.scheduler_execution.arn
  }
}
