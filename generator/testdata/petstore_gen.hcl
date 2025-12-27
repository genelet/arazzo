provider {
  name       = "petStoreDescription"
  server_url = "https://petstore3.swagger.io/api/v3"
  appendices = {
    info_title       = "A pet purchasing workflow"
    info_description = "This workflow walks you through the steps of `searching` for, `selecting`, and `purchasing` an available pet.\n"
    info_summary     = "This workflow showcases how to purchase a pet through a sequence of API calls"
    info_version     = "1.0.1"
  }
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

  outputs {
    available = "$steps.getPetStep.outputs.availablePets"
  }
  
  step "loginStep" {
    description    = "This step demonstrates the user login step"
    operation_id   = "$sourceDescriptions.petStoreDescription.loginUser"
    
    # Mode 2: Parameters as blocks (Commented out due to HCL limitation)
    # parameter {
    #   name  = "username"
    #   in    = "query"
    #   value = "$inputs.username"
    # }
    # parameter {
    #   name  = "password"
    #   in    = "query"
    #   value = "$inputs.password"
    # }

    outputs {
      rateLimit    = "$response.header.X-Rate-Limit"
      sessionToken = "$response.body"
      tokenExpires = "$response.header.X-Expires-After"
    }

    success_criterion {
      condition = "$statusCode == 200"
    }
  }

  step "getPetStep" {
    description    = "retrieve a pet by status from the GET pets endpoint"
    operation_path = "{$sourceDescriptions.petStoreDescription.url}#/paths/~1pet~1findByStatus"

    # parameter {
    #   name  = "status"
    #   in    = "query"
    #   value = "available"
    # }
    # parameter {
    #   name  = "Authorization"
    #   in    = "header"
    #   value = "$steps.loginStep.outputs.sessionToken"
    # }

    outputs {
      availablePets = "$response.body"
    }

    success_criterion {
      condition = "$statusCode == 200"
    }
  }

  step "placeOrderStep" {
    description  = "Place an order for a pet"
    operation_id = "$sourceDescriptions.petStoreDescription.placeOrder"

    request_body = {
      // Simple Payload mode for HCL
      petId    = 1
      quantity = 1
      status   = "placed"
      complete = true
    }

    outputs {
      orderId = "$response.body.id"
    }

    success_criterion {
      condition = "$statusCode == 200"
    }
  }
}
