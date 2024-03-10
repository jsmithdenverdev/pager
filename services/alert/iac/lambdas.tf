module "create_alert_function" {
  source = "./modules/lambda"

  name   = "create_alert"
  path   = "${path.module}/../functions/create-alert"
  bucket = aws_s3_bucket.lambda_bucket.id
}