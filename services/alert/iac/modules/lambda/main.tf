data "archive_file" "lambda" {
  type = "zip"

  source_dir  = "${var.path}/build"
  output_path = "${var.path}/build.zip"
}

resource "aws_s3_object" "lambda" {
  bucket = var.bucket

  key    = "${var.name}.zip"
  source = data.archive_file.lambda.output_path

  etag = filemd5(data.archive_file.lambda.output_path)
}

resource "aws_lambda_function" "function" {
  function_name = var.name

  s3_bucket = var.bucket
  s3_key    = aws_s3_object.lambda.key

  runtime = "provided.al2023"
  handler = "bootstrap"

  source_code_hash = data.archive_file.lambda.output_base64sha256

  role = aws_iam_role.lambda_exec.arn
}

resource "aws_cloudwatch_log_group" "function" {
  name = "/aws/lambda/${aws_lambda_function.function.function_name}"

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