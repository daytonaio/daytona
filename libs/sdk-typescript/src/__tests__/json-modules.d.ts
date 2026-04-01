declare module '*.json' {
  const value: Record<string, unknown>
  export = value
}

declare module '../package.json' {
  const value: { version: string }
  export = value
}
