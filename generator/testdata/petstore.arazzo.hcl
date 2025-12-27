arazzo = "1.0.0"

info {
  title       = "A pet purchasing workflow"
  summary     = "This workflow showcases how to purchase a pet through a sequence of API calls"
  description = "This workflow walks you through the steps of `searching` for, `selecting`, and `purchasing` an available pet.\n"
  version     = "1.0.1"
}

sourceDescriptions "petStoreDescription" {
  url  = "https://raw.githubusercontent.com/swagger-api/swagger-petstore/master/src/main/resources/openapi.yaml"
  type = "openapi"
}

workflow "loginUserRetrievePet" {
  summary     = "Login User and then retrieve pets"
  description = "This procedure lays out the steps to login a user and then retrieve pets"

  inputs {
    type = "object"
    properties = {
      username = { type = "string" }
      password = { type = "string" }
    }
  }

  step "loginStep" {
    description  = "This step demonstrates the user login step"
    operationId = "$sourceDescriptions.petStoreDescription.loginUser"

    parameter {
      name  = "username"
      in    = "query"
      value = "$inputs.username"
    }
    parameter {
      name  = "password"
      in    = "query"
      value = "$inputs.password"
    }

    successCriteria {
      condition = "$statusCode == 200"
    }

    outputs {
      tokenExpires = "$response.header.X-Expires-After"
      rateLimit    = "$response.header.X-Rate-Limit"
      sessionToken = "$response.body"
    }
  }

  step "getPetStep" {
    description   = "retrieve a pet by status from the GET pets endpoint"
    operationPath = "{$sourceDescriptions.petStoreDescription.url}#/paths/~1pet~1findByStatus"

    parameter {
      name  = "status"
      in    = "query"
      value = "available"
    }
    parameter {
      name  = "Authorization"
      in    = "header"
      value = "$steps.loginStep.outputs.sessionToken"
    }

    successCriteria {
      condition = "$statusCode == 200"
    }

    outputs {
      availablePets = "$response.body"
    }
  }

  step "placeOrderStep" {
    description = "Place an order for a pet"
    operationId = "$sourceDescriptions.petStoreDescription.placeOrder"

    requestBody {
      contentType = "application/json"
      payload = {
        petId    = 1
        quantity = 1
        status   = "placed"
        complete = true
      }
    }

    successCriteria {
      condition = "$statusCode == 200"
    }

    outputs {
      orderId = "$response.body.id"
    }
  }

  outputs {
    available = "$steps.getPetStep.outputs.availablePets"
  }
}
