  arazzo = "1.0.0"
  info {
    title       = "Example OAuth service"
    description = "Example OAuth service"
    version     = "1.0.0"
  }
  sourceDescription "apim-auth" {
    url  = "./oauth.openapi.yaml"
    type = "openapi"
  }
  workflow "refresh-token-flow" {
    summary     = "Refresh an access token"
    description = "This is how you can refresh an access token."
    outputs = {
      access_token  = "$steps.do-the-refresh.outputs.access_token"
      expires_in    = "$steps.do-the-refresh.outputs.expires_in"
      refresh_token = "$steps.do-the-refresh.outputs.refresh_token"
    }
    inputs {
      properties "my_client_id" {
        description = "The client id"
        type = "string"
      }
      properties "my_client_secret" {
        type = "string"
        description = "The client secret"
      }
      properties "my_redirect_uri" {
        description = "The redirect uri"
        type = "string"
      }
      type = "object"
    }
    step "do-the-auth-flow" {
      description = "This is where you do the authorization code flow"
      workflowId  = "authorization-code-flow"
      outputs = {
        my_refresh_token = "$outputs.refresh_token"
      }
      parameters = [
        {
          name = "client_id"
          value = "$inputs.my_client_id"
        },
        {
          name = "redirect_uri"
          value = "$inputs.my_redirect_uri"
        },
        {
          name = "client_secret"
          value = "$inputs.my_client_secret"
        }
      ]
    }
    step "do-the-refresh" {
      description = "This is where you do the refresh"
      operationId = "get-token"
      outputs = {
        access_token  = "$response.body#/access_token"
        expires_in    = "$response.body#/expires_in"
        refresh_token = "$response.body#/refresh_token"
      }
      requestBody {
        contentType = "application/x-www-form-urlencoded"
        payload {
          grant_type = "refresh_token"
          refresh_token = "$steps.do-the-auth-flow.outputs.my_refresh_token"
        }
      }
      successCriterion {
        condition = "$statusCode == 200"
      }
      successCriterion {
        context   = "$response.body"
        condition = "$.access_token != null"
        type      = "jsonpath"
      }
    }
  }
  workflow "client-credentials-flow" {
    summary     = "Get an access token using client credentials"
    description = "This is how you can get an access token using client credentials."
    outputs = {
      access_token = "$steps.get-client-creds-token.outputs.access_token"
    }
    inputs {
      type = "object"
      properties "client_id" {
        description = "The client id"
        type = "string"
      }
      properties "client_secret" {
        type = "string"
        description = "The client secret"
      }
    }
    step "get-client-creds-token" {
      description = "This is where you get the token"
      operationId = "get-token"
      outputs = {
        access_token = "$response.body#/access_token"
      }
      requestBody {
        contentType = "application/x-www-form-urlencoded"
        payload {
          client_id = "$inputs.client_id"
          client_secret = "$inputs.client_secret"
          grant_type = "client_credentials"
        }
      }
      successCriterion {
        condition = "$statusCode == 200"
      }
      successCriterion {
        context   = "$response.body"
        condition = "$.access_token != null"
        type      = "jsonpath"
      }
    }
  }
  workflow "authorization-code-flow" {
    summary     = "Get an access token using an authorization code"
    description = "This is how you can get an access token using an authorization code."
    outputs = {
      access_token  = "$steps.get-access-token.outputs.access_token"
      expires_in    = "$steps.get-access-token.outputs.expires_in"
      refresh_token = "$steps.get-access-token.outputs.refresh_token"
    }
    inputs {
      properties "client_id" {
        description = "The client id"
        type = "string"
      }
      properties "client_secret" {
        description = "The client secret"
        type = "string"
      }
      properties "redirect_uri" {
        description = "The redirect uri"
        type = "string"
      }
      type = "object"
    }
    step "browser-authorize" {
      description = "This URL is opened in the browser and redirects you back to the registered redirect URI with an authorization code."
      operationId = "authorize"
      outputs = {
        code = "$response.body#/code"
      }
      parameters = [
        {
          value = "$inputs.client_id"
          in = "query"
          name = "client_id"
        },
        {
          in = "query"
          name = "redirect_uri"
          value = "$inputs.redirect_uri"
        },
        {
          name = "response_type"
          value = "code"
          in = "query"
        },
        {
          in = "query"
          name = "scope"
          value = "read"
        },
        {
          in = "query"
          name = "state"
          value = "12345"
        }
      ]
      successCriterion {
        condition = "$statusCode == 200"
      }
      successCriterion {
        context   = "$response.body"
        condition = "$.access_token != null"
        type      = "jsonpath"
      }
    }
    step "get-access-token" {
      description = "This is where you get the token"
      operationId = "get-token"
      outputs = {
        access_token  = "$response.body#/access_token"
        expires_in    = "$response.body#/expires_in"
        refresh_token = "$response.body#/refresh_token"
      }
      requestBody {
        contentType = "application/x-www-form-urlencoded"
        payload {
          code = "$steps.browser-authorize.outputs.code"
          grant_type = "authorization_code"
          redirect_uri = "$inputs.redirect_uri"
          client_id = "$inputs.client_id"
          client_secret = "$inputs.client_secret"
        }
      }
      successCriterion {
        condition = "$statusCode == 200"
      }
      successCriterion {
        context   = "$response.body"
        condition = "$.access_token != null"
        type      = "jsonpath"
      }
    }
  }