import baseConfig from '../../eslint.config.mjs'

export default [
  ...baseConfig,
  {
    files: ['**/*.ts'],
    rules: {
      'no-restricted-syntax': [
        'error',
        {
          selector:
            'Decorator[expression.callee.name="InjectRepository"] > CallExpression > Identifier[name="Sandbox"]',
          message: 'Do not use @InjectRepository(Sandbox). Use the custom SandboxRepository instead.',
        },
      ],
    },
  },
]
