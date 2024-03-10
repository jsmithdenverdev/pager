data "archive_file" "lambda_alert" {
  type = "zip"

  source_dir  = "${path.module}/../functions/alert/build"
  output_path = "${path.module}/../functions/alert/alert.zip"
}

resource "aws_s3_object" "lambda_alert" {
  bucket = aws_s3_bucket.lambda_bucket.id

  key    = "alert.zip"
  source = data.archive_file.lambda_alert.output_path

  etag = filemd5(data.archive_file.lambda_alert.output_path)
}

resource "aws_lambda_function" "alert" {
  function_name = "alert"

  s3_bucket = aws_s3_bucket.lambda_bucket.id
  s3_key    = aws_s3_object.lambda_alert.key

  runtime = "provided.al2023"
  handler = "bootstrap"

  source_code_hash = data.archive_file.lambda_alert.output_base64sha256

  role = aws_iam_role.lambda_exec.arn
}

resource "aws_cloudwatch_log_group" "alert" {
  name = "/aws/lambda/${aws_lambda_function.alert.function_name}"

  retention_in_days = 30
}

resource "aws_iam_role" "lambda_exec" {
  name = "serverless_lambda"

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Sid    = ""
      Principal = {
        Service = "lambda.amazonaws.com"
      }
      }
    ]
  })
}

resource "aws_iam_role_policy_attachment" "lambda_policy" {
  role       = aws_iam_role.lambda_exec.name
  policy_arn = "arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"
}