plugins {
    application
}

group = "io.daytona.examples"
version = "0.1.0"

java {
    sourceCompatibility = JavaVersion.VERSION_11
    targetCompatibility = JavaVersion.VERSION_11
}

repositories {
    mavenLocal()
    mavenCentral()
}

dependencies {
    implementation("io.daytona:sdk-java")
}

application {
    mainClass.set("io.daytona.examples.DeclarativeImage")
}
