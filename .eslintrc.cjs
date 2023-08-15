module.exports = {
  root: true,
  parser: "@typescript-eslint/parser",
  extends: ["plugin:svelte/recommended"],
  plugins: ["@typescript-eslint"],
  ignorePatterns: ["*.cjs", "*.config.js"],
  overrides: [
    {
      files: ["*.svelte"],
      parser: "svelte-eslint-parser",
      parserOptions: {
        parser: "@typescript-eslint/parser"
      }
    }
  ],
  parserOptions: {
    project: "./tsconfig.json",
    extraFileExtensions: [".svelte"],
    sourceType: "module",
    ecmaVersion: 2020
  },
  env: {
    browser: true,
    es2017: true,
    node: false
  }
};
