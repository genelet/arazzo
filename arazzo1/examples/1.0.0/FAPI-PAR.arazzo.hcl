  arazzo = "1.0.0"
  info {
    title       = "PAR, Authorization and Token workflow"
    description = "A workflow describing how to obtain a token from an OAuth2 and OpenID Connect Financial Grade authorization server which can be common for PSD2 API journeys"
    version     = "1.0.0"
  }
  sourceDescription "auth-api" {
    url  = "./FAPI-PAR.openapi.yaml"
    type = "openapi"
  }
  workflow "OIDC-PAR-AuthzCode" {
    summary     = "PAR, Authorization and Token workflow"
    description = "PAR - Pushed Authorization Request - API Call - https://www.rfc-editor.org/rfc/rfc9126.html Authorize - A web interaction that needs to be passed to a user agent (such as a browser) https://openid.net/specs/openid-connect-core-1_0.html Token - An API call requesting the tokens"
    outputs = {
      access_token = "$steps.TokenStep.outputs.tokenResponse"
    }
    inputs {
      required = [
        "PARrequestBody",
        "TokenRequestBody"
      ]
      type = "object"
      properties "PARrequestBody" {
        description = "Parameters that comprise an authorization request are sent directly to the 
pushed authorization request endpoint in the request body
[PAR Request](https://tools.ietf.org/html/draft-ietf-oauth-par-07#section-2.1)            
"
        properties "client_id" {
          type = "string"
        }
        properties "code_challenge" {
          type = "string"
        }
        properties "code_challenge_method" {
          type = "string"
        }
        properties "consent_id" {
          type = "string"
        }
        properties "nonce" {
          type = "string"
        }
        properties "prompt" {
          type = "string"
        }
        properties "redirect_uri" {
          type = "string"
        }
        properties "response_type" {
          type = "string"
        }
        properties "scope" {
          type = "string"
        }
        properties "state" {
          type = "string"
        }
        properties "sub" {
          type = "string"
        }
        type = "object"
      }
      properties "TokenRequestBody" {
        properties "code" {
          type = "string"
        }
        properties "code_verifier" {
          type = "string"
        }
        properties "grant_type" {
          type = "string"
        }
        properties "redirect_uri" {
          type = "string"
        }
        required = [
          "grant_type",
          "code",
          "redirect_uri",
          "code_verifier"
        ]
        type = "object"
        description = "Request Schema for the token endpoint in the context of an OAuth2 Authorization code flow (**Note** this is place holder object that will have values replaced dynamically)"
      }
      properties "client_assertion" {
        properties "aud" {
          type = "string"
        }
        properties "exp" {
          type = "string"
        }
        properties "iat" {
          type = "string"
        }
        properties "iss" {
          type = "string"
        }
        properties "jti" {
          type = "string"
        }
        properties "sub" {
          type = "string"
        }
        type = "object"
        description = "Used for PAR client authentication. The assertion contains a JWS, in this an object `base64(JWS)` 
signed with JWT signing private key related to the TPP OAuth client. See the Model and the Assertion 
object for a detailed description of the content.           
"
      }
      properties "client_id" {
        description = "The identifier of the third party provider OAuth client. ClientId is returned during the TPP registration."
        type = "string"
      }
      properties "code_verifier" {
        description = "The code verifier Proof Key of Code Exchange (PKCE)"
        type = "string"
      }
      properties "redirect_uri" {
        description = "The value of the redirect URI that was used in the previous `/as/authorize.oauth2` call."
        type = "string"
      }
    }
    step "PARStep" {
      description = "Pushed Authorization Request"
      operationId = "$sourceDescriptions.auth-api.PAR"
      outputs = {
        request_uri = "$response.body#/request_uri"
      }
      parameters = [
        {
          in = "query"
          name = "client_id"
          value = "$inputs.client_id"
        },
        {
          in = "query"
          name = "client_assertion_type"
          value = "urn:ietf:params:oauth:grant-type:jwt-bearer"
        },
        {
          in = "query"
          name = "client_assertion"
          value = "$inputs.client_assertion"
        }
      ]
      requestBody {
        payload "$inputs.PARrequestBody"
      }
      successCriterion {
        condition = "$statusCode == 200"
      }
    }
    step "AuthzCodeStep" {
      description = "OIDC Authorization code request"
      operationId = "$sourceDescriptions.auth-api.Authorization"
      outputs = {
        code = "$response.body#/code"
      }
      parameters = [
        {
          in = "query"
          name = "request_uri"
          value = "$steps.PARStep.outputs.request_uri"
        },
        {
          in = "query"
          name = "client_id"
          value = "$inputs.client_id"
        }
      ]
      successCriterion {
        condition = "$statusCode == 302"
      }
    }
    step "TokenStep" {
      description = "Get token from the OIDC Token endpoint"
      operationId = "$sourceDescriptions.auth-api.Token"
      outputs = {
        tokenResponse = "$response.body"
      }
      parameters = [
        {
          in = "query"
          name = "client_id"
          value = "$inputs.client_id"
        },
        {
          in = "query"
          name = "client_assertion_type"
          value = "urn:ietf:params:oauth:grant-type:jwt-bearer"
        },
        {
          in = "query"
          name = "client_assertion"
          value = "$inputs.client_assertion"
        }
      ]
      requestBody {
        payload "{
  "grant_type": "authorization_code",
  "code": "{$steps.AuthzCodeStep.outputs.code}",
  "redirect_uri": "{$inputs.redirect_uri}",
  "code_verifier": "{$inputs.code_verifier}"
}                      
"
      }
      successCriterion {
        condition = "$statusCode == 200"
      }
    }
  }