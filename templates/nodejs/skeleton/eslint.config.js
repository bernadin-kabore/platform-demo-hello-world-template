module.exports = [
  {
    languageOptions: {
      ecmaVersion: 2022,
      sourceType: "commonjs",
      globals: { require: "readonly", module: "readonly", process: "readonly", console: "readonly" },
    },
    rules: {
      "no-unused-vars": "warn",
      "no-undef": "error",
    },
  },
];
