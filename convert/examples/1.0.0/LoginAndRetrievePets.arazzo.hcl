  arazzo = "1.0.0"
  info {
    title       = "A pet purchasing workflow"
    summary     = "This workflow showcases how to purchase a pet through a sequence of API calls"
    description = "This workflow walks you through the steps of `searching` for, `selecting`, and `purchasing` an available pet.\\n"
    version     = "1.0.1"
  }
  sourceDescription "petStoreDescription" {
    url  = "https://raw.githubusercontent.com/swagger-api/swagger-petstore/master/src/main/resources/openapi.yaml"
    type = "openapi"
  }
  workflow "loginUserRetrievePet" {
    summary     = "Login User and then retrieve pets"
    description = "This procedure lays out the steps to login a user and then retrieve pets"
    outputs = {
      available = "$steps.getPetStep.outputs.availablePets"
    }
    inputs {
      properties "password" {
        type = "string"
      }
      properties "username" {
        type = "string"
      }
      type = "object"
    }
    step "loginStep" {
      description = "This step demonstrates the user login step"
      operationId = "$sourceDescriptions.petStoreDescription.loginUser"
      outputs = {
        rateLimit    = "$response.header.X-Rate-Limit"
        sessionToken = "$response.body"
        tokenExpires = "$response.header.X-Expires-After"
      }
      parameters = [
        {
          in = "query"
          name = "username"
          value = "$inputs.username"
        },
        {
          in = "query"
          name = "password"
          value = "$inputs.password"
        }
      ]
      successCriterion {
        condition = "$statusCode == 200"
      }
    }
    step "getPetStep" {
      description   = "retrieve a pet by status from the GET pets endpoint"
      operationPath = "{$sourceDescriptions.petStoreDescription.url}#/paths/~1pet~1findByStatus"
      outputs = {
        availablePets = "$response.body"
      }
      parameters = [
        {
          in = "query"
          name = "status"
          value = "available"
        },
        {
          in = "header"
          name = "Authorization"
          value = "$steps.loginStep.outputs.sessionToken"
        }
      ]
      successCriterion {
        condition = "$statusCode == 200"
      }
    }
  }