// @ts-check
import eslintPluginPrettier from 'eslint-plugin-prettier';
import eslintPluginSimpleImportSort from 'eslint-plugin-simple-import-sort';
import eslintPluginUnusedImports from 'eslint-plugin-unused-imports';

import withNuxt from './.nuxt/eslint.config.mjs';

export default withNuxt([
  { ignores: ['dist', 'build', '.nuxt', '.output', 'node_modules'] },
  {
    files: ['**/*.{ts,vue,js,mjs}'],
    plugins: {
      prettier: eslintPluginPrettier,
      'unused-imports': eslintPluginUnusedImports,
      'simple-import-sort': eslintPluginSimpleImportSort,
    },
    rules: {
      'unused-imports/no-unused-imports': 'error',
      'simple-import-sort/imports': 'error',
      'simple-import-sort/exports': 'error',
      'prettier/prettier': 'error',
      eqeqeq: ['error', 'always', { null: 'ignore' }],
      'no-console': ['warn', { allow: ['warn', 'error'] }],
      'no-debugger': 'error',
      'prefer-const': 'error',
    },
  },
]);
