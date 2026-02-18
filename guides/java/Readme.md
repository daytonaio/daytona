# Daytona Integration with Java

This guide shows how to interact with the Daytona API using Java. Since there is no official Java SDK at the moment, we interact with Daytona’s REST API using the standard Java HTTP client (Java 11+).

## Prerequisites

- Java 11 or newer
- A Daytona API key
- Maven or Gradle (optional, for dependency management)

## Authenticate with Daytona

Before making API requests, ensure you have a Daytona API key. Store it as an environment variable:

```bash
export DAYTONA_API_KEY="your_api_key_here"