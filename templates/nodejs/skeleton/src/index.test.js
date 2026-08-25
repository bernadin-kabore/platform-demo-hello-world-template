const test = require("node:test");
const assert = require("node:assert");
const http = require("node:http");

// Minimal smoke test the CI pipeline's "test" stage runs. Real services
// substitute their own test suite; this just proves the pipeline plumbing
// (test -> SAST -> SCA -> build -> scan -> sign -> SBOM) works end to end.
test("index module loads without throwing", () => {
  assert.doesNotThrow(() => require("./index.js"));
});
