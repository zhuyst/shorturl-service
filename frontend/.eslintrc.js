module.exports = {
    root: true,
    parserOptions: {
        parser: 'babel-eslint',
        ecmaVersion: 2019,
        sourceType: 'module'
    },
    env: {
        es6: true,
        browser: true
    },
    extends: [
      'eslint-config-egg',
    ],
    plugins: [
        'svelte3'
    ],
    overrides: [
        {
          files: ['**/*.svelte'],
          processor: 'svelte3/svelte3'
        }
    ],
    rules: {}
};