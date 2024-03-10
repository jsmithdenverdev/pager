resource "aws_apigatewayv2_integration" "function" {
  api_id = var.api_id

  integration_uri    = var.integration_uri
  integration_type   = "AWS_PROXY"
  integration_method = "POST"
}

resource "aws_apigatewayv2_route" "route" {
  api_id = var.api_id

  route_key = "${var.method} ${var.path}"
  target    = "integrations/${aws_apigatewayv2_integration.function.id}"
}

resource "aws_lambda_permission" "api_gw" {
  statement_id  = "AllowExecutionFromAPIGateway"
  action        = "lambda:InvokeFunction"
  function_name = var.function_name
  principal     = "apigateway.amazonaws.com"
  source_arn    = var.source_arn
}
