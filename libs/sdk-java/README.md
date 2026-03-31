# Daytona Java SDK

Java SDK for Daytona APIs (Java 11+), using OkHttp and Jackson.

## Configuration

Environment variables:

- `DAYTONA_API_KEY`
- `DAYTONA_API_URL` (default: `https://app.daytona.io/api`)
- `DAYTONA_TARGET`

## Quick start

```java
Daytona daytona = new Daytona();
Sandbox sandbox = daytona.create();
ExecuteResponse response = sandbox.getProcess().executeCommand("echo hello");
System.out.println(response.getResult());
sandbox.delete();
daytona.close();
```
