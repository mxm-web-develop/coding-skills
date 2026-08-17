import assert from "node:assert/strict";
import test from "node:test";

import { sellerStatusLabel } from "../src/seller-status.js";

test("uses a readable label for review status", () => {
  assert.equal(sellerStatusLabel("review"), "待审核");
});
