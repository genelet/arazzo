  arazzo = "1.0.0"
  info {
    title   = "Public Zoo API"
    version = "1.0"
  }
  sourceDescription "animals" {
    url  = "./animals.yaml"
    type = "openapi"
  }
  workflow "animal-workflow" {
    step "post-step" {
      operationId = "$sourceDescriptions.animals.postAnimal"
      parameters = [
        {
          name = "authentication"
          value = "SUPER_SECRET"
          in = "cookie"
        }
      ]
    }
    step "get-step" {
      operationId = "$sourceDescriptions.animals.getRandomAnimal"
    }
  }