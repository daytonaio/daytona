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

    testImplementation("org.junit.jupiter:junit-jupiter:5.11.4")
    testImplementation("org.assertj:assertj-core:3.26.3")
    testImplementation("org.mockito:mockito-core:5.14.2")
    testImplementation("org.mockito:mockito-junit-jupiter:5.14.2")
    testImplementation("com.squareup.okhttp3:mockwebserver:4.12.0")
}

tasks.test {
    useJUnitPlatform()
    exclude("**/E2ETest.class")
    jvmArgs(
        "--add-opens=java.base/java.lang=ALL-UNNAMED",
        "--add-opens=java.base/java.util=ALL-UNNAMED"
    )
}

tasks.register<Test>("testE2E") {
    description = "Runs the end-to-end test suite (requires DAYTONA_API_KEY)."
    group = "verification"
    useJUnitPlatform()
    include("**/E2ETest.class")
    jvmArgs(
        "--add-opens=java.base/java.lang=ALL-UNNAMED",
        "--add-opens=java.base/java.util=ALL-UNNAMED"
    )
    testLogging {
        events("passed", "skipped", "failed")
        showStandardStreams = true
    }
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
