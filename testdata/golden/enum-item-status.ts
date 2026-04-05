import type { ApiEnum } from "./types";

const enumData: ApiEnum = {
  "name": "ItemStatus",
  "package": "testpkg.v1",
  "description": "ItemStatus represents the status of an item.",
  "values": [
    {
      "name": "ITEM_STATUS_UNSPECIFIED",
      "number": 0,
      "description": "Unspecified status."
    },
    {
      "name": "ITEM_STATUS_ACTIVE",
      "number": 1,
      "description": "Item is active."
    },
    {
      "name": "ITEM_STATUS_ARCHIVED",
      "number": 2,
      "description": "Item is archived."
    }
  ],
  "usedBy": [
    "ItemService"
  ]
};

export default enumData;
