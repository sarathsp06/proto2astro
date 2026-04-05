import type { ApiService } from "./types";

const service: ApiService = {
  "service": "ItemService",
  "package": "testpkg.v1",
  "description": "ItemService manages items.",
  "rpcs": [
    {
      "name": "CreateItem",
      "description": "CreateItem creates a new item.",
      "request": [
        {
          "name": "name",
          "type": "string",
          "required": true,
          "description": "The name of the item.",
          "example": "My Item"
        },
        {
          "name": "count",
          "type": "int32",
          "description": "A description with HTML chars: values < 100 & > 0.",
          "example": 50
        },
        {
          "name": "priority",
          "type": "Priority",
          "description": "Priority level for the item."
        },
        {
          "name": "tags",
          "type": "string[]",
          "description": "Optional tags for the item.",
          "example": [
            "alpha",
            "beta"
          ]
        }
      ],
      "response": [
        {
          "name": "item",
          "type": "ItemDetail",
          "description": "The created item."
        }
      ],
      "errors": [
        {
          "code": "ALREADY_EXISTS",
          "description": "An item with the same name exists."
        },
        {
          "code": "INVALID_ARGUMENT",
          "description": "The name is empty."
        }
      ]
    }
  ]
};

export default service;
