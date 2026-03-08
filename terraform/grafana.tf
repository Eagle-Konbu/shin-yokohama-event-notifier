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

# -----------------------------------------------------------------------------
# Grafana Dashboard
# -----------------------------------------------------------------------------

resource "grafana_dashboard" "lambda" {
  config_json = jsonencode({
    title       = "Lambda: ${var.project_name}"
    description = "Monitoring dashboard for ${local.function_name}"
    editable    = false
    timezone    = "Asia/Tokyo"

    time = {
      from = "now-7d"
      to   = "now"
    }
    refresh = "1h"

    panels = [
      {
        id    = 1
        title = "Invocations"
        type  = "timeseries"
        gridPos = {
          h = 8
          w = 12
          x = 0
          y = 0
        }
        datasource = {
          type = "cloudwatch"
          uid  = grafana_data_source.cloudwatch.uid
        }
        targets = [
          {
            refId      = "A"
            namespace  = "AWS/Lambda"
            metricName = "Invocations"
            dimensions = {
              FunctionName = [local.function_name]
            }
            statistic = "Sum"
            period    = "86400"
            region    = var.aws_region
          }
        ]
        fieldConfig = {
          defaults = {
            color = {
              mode = "palette-classic"
            }
            custom = {
              drawStyle   = "bars"
              fillOpacity = 50
            }
          }
        }
      },
      {
        id    = 2
        title = "Errors"
        type  = "timeseries"
        gridPos = {
          h = 8
          w = 12
          x = 12
          y = 0
        }
        datasource = {
          type = "cloudwatch"
          uid  = grafana_data_source.cloudwatch.uid
        }
        targets = [
          {
            refId      = "A"
            namespace  = "AWS/Lambda"
            metricName = "Errors"
            dimensions = {
              FunctionName = [local.function_name]
            }
            statistic = "Sum"
            period    = "86400"
            region    = var.aws_region
          }
        ]
        fieldConfig = {
          defaults = {
            color = {
              fixedColor = "red"
              mode       = "fixed"
            }
            custom = {
              drawStyle   = "bars"
              fillOpacity = 50
            }
          }
        }
      },
      {
        id    = 3
        title = "Error Rate (%)"
        type  = "stat"
        gridPos = {
          h = 8
          w = 6
          x = 0
          y = 8
        }
        datasource = {
          type = "cloudwatch"
          uid  = grafana_data_source.cloudwatch.uid
        }
        targets = [
          {
            refId      = "errors"
            namespace  = "AWS/Lambda"
            metricName = "Errors"
            dimensions = {
              FunctionName = [local.function_name]
            }
            statistic = "Sum"
            period    = "86400"
            region    = var.aws_region
            id        = "errors"
            hide      = true
          },
          {
            refId      = "invocations"
            namespace  = "AWS/Lambda"
            metricName = "Invocations"
            dimensions = {
              FunctionName = [local.function_name]
            }
            statistic = "Sum"
            period    = "86400"
            region    = var.aws_region
            id        = "invocations"
            hide      = true
          },
          {
            refId      = "C"
            type       = "math"
            expression = "errors / invocations * 100"
          }
        ]
        fieldConfig = {
          defaults = {
            unit = "percent"
            thresholds = {
              mode = "absolute"
              steps = [
                { color = "green", value = null },
                { color = "yellow", value = 10 },
                { color = "red", value = 50 }
              ]
            }
          }
        }
        options = {
          reduceOptions = {
            calcs = ["lastNotNull"]
          }
          colorMode = "background"
        }
      },
      {
        id    = 4
        title = "Duration"
        type  = "timeseries"
        gridPos = {
          h = 8
          w = 18
          x = 6
          y = 8
        }
        datasource = {
          type = "cloudwatch"
          uid  = grafana_data_source.cloudwatch.uid
        }
        targets = [
          {
            refId      = "A"
            namespace  = "AWS/Lambda"
            metricName = "Duration"
            dimensions = {
              FunctionName = [local.function_name]
            }
            statistic = "Average"
            period    = "86400"
            region    = var.aws_region
            label     = "Average"
          },
          {
            refId      = "B"
            namespace  = "AWS/Lambda"
            metricName = "Duration"
            dimensions = {
              FunctionName = [local.function_name]
            }
            statistic = "Maximum"
            period    = "86400"
            region    = var.aws_region
            label     = "Maximum"
          },
          {
            refId      = "C"
            namespace  = "AWS/Lambda"
            metricName = "Duration"
            dimensions = {
              FunctionName = [local.function_name]
            }
            statistic = "p99"
            period    = "86400"
            region    = var.aws_region
            label     = "p99"
          }
        ]
        fieldConfig = {
          defaults = {
            unit = "ms"
            color = {
              mode = "palette-classic"
            }
          }
        }
      },
      {
        id    = 5
        title = "Throttles"
        type  = "timeseries"
        gridPos = {
          h = 8
          w = 12
          x = 0
          y = 16
        }
        datasource = {
          type = "cloudwatch"
          uid  = grafana_data_source.cloudwatch.uid
        }
        targets = [
          {
            refId      = "A"
            namespace  = "AWS/Lambda"
            metricName = "Throttles"
            dimensions = {
              FunctionName = [local.function_name]
            }
            statistic = "Sum"
            period    = "86400"
            region    = var.aws_region
          }
        ]
        fieldConfig = {
          defaults = {
            color = {
              fixedColor = "orange"
              mode       = "fixed"
            }
            custom = {
              drawStyle   = "bars"
              fillOpacity = 50
            }
          }
        }
      },
      {
        id    = 6
        title = "ConcurrentExecutions"
        type  = "timeseries"
        gridPos = {
          h = 8
          w = 12
          x = 12
          y = 16
        }
        datasource = {
          type = "cloudwatch"
          uid  = grafana_data_source.cloudwatch.uid
        }
        targets = [
          {
            refId      = "A"
            namespace  = "AWS/Lambda"
            metricName = "ConcurrentExecutions"
            dimensions = {
              FunctionName = [local.function_name]
            }
            statistic = "Maximum"
            period    = "86400"
            region    = var.aws_region
          }
        ]
        fieldConfig = {
          defaults = {
            color = {
              mode = "palette-classic"
            }
          }
        }
      }
    ]
  })
}
