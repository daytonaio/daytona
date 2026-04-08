# Daytona API Client for Java

Auto-generated Java client for the [Daytona](https://daytona.io) REST API. This library is used internally by the [Daytona Java SDK](https://central.sonatype.com/artifact/io.daytona/sdk) and is not intended for direct use.

## Usage

If you're building applications with Daytona, use the [Daytona Java SDK](https://central.sonatype.com/artifact/io.daytona/sdk) instead — it provides a higher-level, idiomatic Java interface.

```kotlin
dependencies {
    implementation("io.daytona:sdk:<version>")
}
```

## Generation

This client is generated from the Daytona OpenAPI specification using [OpenAPI Generator](https://openapi-generator.tech):

```bash
yarn nx run api-client-java:generate:api-client
```

Do not edit the generated source files manually — changes will be overwritten on regeneration.

## License

Apache License 2.0 — see [LICENSE](https://github.com/daytonaio/daytona/blob/main/libs/sdk-java/LICENSE) for details.
