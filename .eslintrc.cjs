/* eslint-env node */
module.exports = {
  env: {
    browser: true,
    node: true,
    es2022: true,
  },
  extends: ["eslint:recommended"],
  overrides: [
    {
      files: ["**/*.test.js"],
      env: { jest: true },
    },
  ],
  ignorePatterns: ["node_modules/", "coverage/"],
  root: true,
};
