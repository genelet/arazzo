provider {
  name       = "sample-provider"
  server_url = "https://api.staging.example.com/v1"
}

workflow "full-demo-workflow" {
  summary     = "Demonstrates all generator features"
  description = "A comprehensive workflow showing parameters, inputs, outputs, and request bodies."

  # Inputs (HCL map/object syntax)
  inputs {
    type       = "object"
    properties = {
      username     = { type = "string" }
      password     = { type = "string" }
      targetUserId = { type = "integer" }
    }
    required   = ["username", "password", "targetUserId"]
  }

  outputs {
    authToken = "$steps.login.outputs.token"
    userName  = "$steps.fetchUser.outputs.name"
  }

  step "login" {
    operation_id = "login"
    # request_body omitted for auto-generation

    outputs {
      token = "$response.body.token"
    }
    
    success_criterion {
      condition = "$statusCode == 200"
    }
  }

  step "fetchUser" {
    operation_id = "getUser"
    
    # Mode 1: Auto-include mandatory 'id' override
    # parameter {
    #   name  = "id"
    #   in    = "path"
    #   value = "$inputs.targetUserId"
    # }
    
    # Mode 2
    # parameter {
    #   name = "verbose"
    # }

    # Mode 3
    # parameter {
    #   name  = "X-Trace-Id"
    #   in    = "header"
    #   value = "trace-12345"
    # }

    outputs {
      name = "$response.body.name"
    }
  }

  step "updateProfile" {
    operation_id = "updateProfile"
    
    # parameter {
    #   name  = "id"
    #   in    = "path"
    #   value = 999
    # }

    # Explicit Simple RequestBody
    # Since RequestBody is interface{}, HCL decoding might be tricky if we want a simple map.
    # But we have `hcl:"request_body,block"`.
    # This implies it expects a block structure. 
    # If we want a simple map payload, we can define attributes inside the block.
    request_body {
      bio     = "Updated bio from generator"
      website = "https://genelet.com"
    }

    on_success {
      name = "logSuccess"
      type = "schema"
    }
    
    on_failure {
      name = "logFailure"
      type = "schema"
    }
  }
}
