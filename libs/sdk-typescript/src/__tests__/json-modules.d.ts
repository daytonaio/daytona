declare module '*.json' {
  const value: Record<string, unknown>
  export = value
}

declare module '../package.json' {
  const value: { name: string; version: string }
  export = value
}
