rootProject.name = "lifecycle"

dependencyResolutionManagement {
    repositories {
        mavenLocal()
        mavenCentral()
    }
}

includeBuild("../../../libs/sdk-java") {
    dependencySubstitution {
        substitute(module("io.daytona:sdk-java")).using(project(":"))
    }
}
