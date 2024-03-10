output invoke_arn {
  description = "The invoke arn of the lambda function"
  value       = aws_lambda_function.function.invoke_arn
}

output function_name {
  description = "The name of the lambda function"
  value       = aws_lambda_function.function.function_name
}