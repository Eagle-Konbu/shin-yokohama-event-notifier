# -----------------------------------------------------------------------------
# Grafana Cloud Integration
# -----------------------------------------------------------------------------

# IAM Policy for Grafana Cloud to access CloudWatch and Cost Explorer
resource "aws_iam_policy" "grafana_cloud" {
  name = "${var.project_name}-grafana-cloud-policy"

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "CloudWatchReadOnly"
        Effect = "Allow"
        Action = [
          "cloudwatch:DescribeAlarmsForMetric",
          "cloudwatch:DescribeAlarmHistory",
          "cloudwatch:DescribeAlarms",
          "cloudwatch:ListMetrics",
          "cloudwatch:GetMetricData",
          "cloudwatch:GetInsightRuleReport"
        ]
        Resource = "*"
      },
      {
        Sid    = "CloudWatchLogsReadOnly"
        Effect = "Allow"
        Action = [
          "logs:DescribeLogGroups",
          "logs:GetLogGroupFields",
          "logs:StartQuery",
          "logs:StopQuery",
          "logs:GetQueryResults",
          "logs:GetLogEvents"
        ]
        Resource = "*"
      },
      {
        Sid    = "CostExplorerReadOnly"
        Effect = "Allow"
        Action = [
          "ce:GetCostAndUsage",
          "ce:GetCostForecast"
        ]
        Resource = "*"
      },
      {
        Sid    = "TagReadOnly"
        Effect = "Allow"
        Action = [
          "tag:GetResources"
        ]
        Resource = "*"
      }
    ]
  })

  tags = local.common_tags
}

# IAM User for Grafana Cloud
resource "aws_iam_user" "grafana_cloud" {
  name = "${var.project_name}-grafana-cloud"
  tags = local.common_tags
}

resource "aws_iam_user_policy_attachment" "grafana_cloud" {
  user       = aws_iam_user.grafana_cloud.name
  policy_arn = aws_iam_policy.grafana_cloud.arn
}

resource "aws_iam_access_key" "grafana_cloud" {
  user = aws_iam_user.grafana_cloud.name
}

# -----------------------------------------------------------------------------
# Grafana Data Source
# -----------------------------------------------------------------------------

resource "grafana_data_source" "cloudwatch" {
  type = "cloudwatch"
  name = "AWS CloudWatch"

  json_data_encoded = jsonencode({
    defaultRegion = var.aws_region
    authType      = "keys"
  })

  secure_json_data_encoded = jsonencode({
    accessKey = aws_iam_access_key.grafana_cloud.id
    secretKey = aws_iam_access_key.grafana_cloud.secret
  })
}
