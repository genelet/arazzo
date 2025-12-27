provider {
  name       = "test-service"
  server_url = "http://api.example.com"
}

http "my-op" {
  service_type = "http"
  path         = "/test"
  method       = "GET"
}
