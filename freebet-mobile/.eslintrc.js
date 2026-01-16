module.exports = {
  root: true,
  extends: [
    'expo',
    'eslint:recommended',
    '@typescript-eslint/recommended',
  ],
  parser: '@typescript-eslint/parser',
  plugins: ['@typescript-eslint'],
  rules: {
    // Add custom rules here
    '@typescript-eslint/no-unused-vars': ['error', { argsIgnorePattern: '^_' }],
    'react-hooks/exhaustive-deps': 'warn',
  },
  env: {
    node: true,
  },
  ignorePatterns: [
    'node_modules/',
    '.expo/',
    'dist/',
  ],
};
