plugins {
    `java-library`
    `maven-publish`
    signing
}

group = "io.daytona"
version = "0.0.0-dev"

val depVersion = version.toString()

java {
    sourceCompatibility = JavaVersion.VERSION_11
    targetCompatibility = JavaVersion.VERSION_11
    withJavadocJar()
    withSourcesJar()
}

repositories {
    mavenLocal()
    mavenCentral()
}

dependencies {
    api("io.daytona:api-client:$depVersion")
    api("io.daytona:toolbox-api-client:$depVersion")
    api("com.squareup.okhttp3:okhttp:4.12.0")
    api("com.fasterxml.jackson.core:jackson-databind:2.17.2")
    api("com.fasterxml.jackson.core:jackson-annotations:2.17.2")
    implementation("io.socket:socket.io-client:2.1.2")
}

publishing {
    publications {
        create<MavenPublication>("mavenJava") {
            artifactId = "sdk"
            from(components["java"])

            pom {
                name.set("Daytona Java SDK")
                description.set("Official Java SDK for Daytona — secure, elastic cloud infrastructure for running AI-generated code")
                url.set("https://github.com/daytonaio/daytona")

                licenses {
                    license {
                        name.set("Apache License, Version 2.0")
                        url.set("https://www.apache.org/licenses/LICENSE-2.0")
                    }
                }

                developers {
                    developer {
                        id.set("daytonaio")
                        name.set("Daytona Platforms Inc.")
                        email.set("support@daytona.io")
                    }
                }

                scm {
                    connection.set("scm:git:git://github.com/daytonaio/daytona.git")
                    developerConnection.set("scm:git:ssh://github.com:daytonaio/daytona.git")
                    url.set("https://github.com/daytonaio/daytona")
                }
            }
        }
    }

    repositories {
        maven {
            name = "release"
            if (version.toString().endsWith("-SNAPSHOT")) {
                url = uri("https://central.sonatype.com/repository/maven-snapshots/")
                credentials {
                    username = System.getenv("MAVEN_USERNAME") ?: ""
                    password = System.getenv("MAVEN_PASSWORD") ?: ""
                }
            } else {
                url = uri(layout.buildDirectory.dir("staging-deploy"))
            }
        }
    }
}

signing {
    val rawKey = System.getenv("MAVEN_GPG_SIGNING_KEY")
    val signingKey = rawKey?.replace("\\n", "\n")
    val signingPassword = System.getenv("MAVEN_GPG_SIGNING_PASSWORD")
    if (!signingKey.isNullOrBlank()) {
        useInMemoryPgpKeys(signingKey, signingPassword ?: "")
        sign(publishing.publications["mavenJava"])
    }
}

tasks.withType<Jar> {
    manifest {
        attributes("Implementation-Version" to project.version)
    }
}

tasks.withType<Javadoc> {
    (options as StandardJavadocDocletOptions).addStringOption("Xdoclint:none", "-quiet")
}
