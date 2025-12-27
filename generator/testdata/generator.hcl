provider {
  name       = "test-service"
  server_url = "http://api.example.com"
}

workflow "main" {
  step "my-op" {
    operation_path = "/test"
    description    = "My Operation"
  }
}
