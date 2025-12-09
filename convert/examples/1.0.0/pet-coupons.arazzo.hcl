  arazzo = "1.0.0"
  info {
    title       = "Petstore - Apply Coupons"
    description = "Illustrates a workflow whereby a client a) finds a pet in the petstore,  b) finds coupons for that pet, and finally c) orders the pet while applying the coupons from step b."
    version     = "1.0.0"
  }
  sourceDescription "pet-coupons" {
    url  = "./pet-coupons.openapi.yaml"
    type = "openapi"
  }
  workflow "apply-coupon" {
    summary     = "Apply a coupon to a pet order."
    description = "This is how you can find a pet, find an applicable coupon, and apply that coupon in your order. The workflow concludes by outputting the ID of the placed order."
    outputs = {
      apply_coupon_pet_order_id = "$steps.place-order.outputs.my_order_id"
    }
    inputs {
      _ref = "#/components/inputs/apply_coupon_input"
    }
    step "find-pet" {
      description = "Find a pet based on the provided tags."
      operationId = "findPetsByTags"
      outputs = {
        my_pet_id = "$response.body#/0/id"
      }
      parameters = [
        {
          in = "query"
          name = "pet_tags"
          value = "$inputs.my_pet_tags"
        }
      ]
      successCriterion {
        condition = "$statusCode == 200"
      }
    }
    step "find-coupons" {
      description = "Find a coupon available for the selected pet."
      operationId = "getPetCoupons"
      outputs = {
        my_coupon_code = "$response.body#/couponCode"
      }
      parameters = [
        {
          in = "path"
          name = "pet_id"
          value = "$steps.find-pet.outputs.my_pet_id"
        }
      ]
      successCriterion {
        condition = "$statusCode == 200"
      }
    }
    step "place-order" {
      description = "Place an order for the pet, applying the coupon."
      workflowId  = "place-order"
      outputs = {
        my_order_id = "$outputs.workflow_order_id"
      }
      parameters = [
        {
          name = "pet_id"
          value = "$steps.find-pet.outputs.my_pet_id"
        },
        {
          name = "coupon_code"
          value = "$steps.find-coupons.outputs.my_coupon_code"
        }
      ]
      successCriterion {
        condition = "$statusCode == 200"
      }
    }
  }
  workflow "buy-available-pet" {
    summary     = "Buy an available pet if one is available."
    description = "This workflow demonstrates a workflow very similar to `apply-coupon`, by intention. It's meant to indicate how to reuse a step (`place-order`) as well as a parameter (`page`, `pageSize`)."
    outputs = {
      buy_pet_order_id = "$steps.place-order.outputs.my_order_id"
    }
    inputs {
      _ref = "#/components/inputs/buy_available_pet_input"
    }
    step "find-pet" {
      description = "Find a pet that is available for purchase."
      operationId = "findPetsByStatus"
      outputs = {
        my_pet_id = "$response.body#/0/id"
      }
      parameters = [
        {
          value = "available"
          in = "query"
          name = "status"
        },
        {
          reference = "$components.parameters.page"
          value = 1
        },
        {
          reference = "$components.parameters.pageSize"
          value = 10
        }
      ]
      successCriterion {
        condition = "$statusCode == 200"
      }
    }
    step "place-order" {
      description = "Place an order for the pet."
      workflowId  = "place-order"
      outputs = {
        my_order_id = "$outputs.workflow_order_id"
      }
      parameters = [
        {
          value = "$steps.find-pet.outputs.my_pet_id"
          name = "pet_id"
        }
      ]
      successCriterion {
        condition = "$statusCode == 200"
      }
    }
  }
  workflow "place-order" {
    summary     = "Place an order for a pet."
    description = "This workflow places an order for a pet. It may be reused by other workflows as the \"final step\" in a purchase."
    outputs = {
      workflow_order_id = "$steps.place-order.outputs.step_order_id"
    }
    inputs {
      properties "coupon_code" {
        description = "The coupon code to apply to the order."
        type = "string"
      }
      properties "pet_id" {
        type = "integer"
        description = "The ID of the pet to place in the order."
        format = "int64"
      }
      properties "quantity" {
        description = "The number of pets to place in the order."
        format = "int32"
        type = "integer"
      }
      type = "object"
    }
    step "place-order" {
      description = "Place an order for the pet."
      operationId = "placeOrder"
      outputs = {
        step_order_id = "$response.body#/id"
      }
      requestBody {
        contentType = "application/json"
        payload {
          status = "placed"
          complete = false
          couponCode = "$inputs.coupon_code"
          petId = "$inputs.pet_id"
          quantity = "$inputs.quantity"
        }
      }
      successCriterion {
        condition = "$statusCode == 200"
      }
    }
  }
  components {
    inputs {
      apply_coupon_input {
        properties "my_pet_tags" {
          description = "Desired tags to use when searching for a pet, in CSV format (e.g. \"puppy, dalmatian\")"
          items {
            type = "string"
          }
          type = "array"
        }
        properties "store_id" {
          _ref = "#/components/inputs/store_id"
        }
        type = "object"
      }
      buy_available_pet_input {
        properties "store_id" {
          _ref = "#/components/inputs/store_id"
        }
        type = "object"
      }
      store_id {
        description = "Indicates the domain name of the store where the customer is browsing or buying pets, e.g. \"pets.example.com\" or \"pets.example.co.uk\"."
        type = "string"
      }
    }
    parameter "page" {
      in = "query"
      value = 1
    }
    parameter "pageSize" {
      in = "query"
      value = 100
    }
  }