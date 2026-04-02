rootProject.name = "sdk-java"

includeBuild("../api-client-java") {
    name = "api-client"
    dependencySubstitution {
        substitute(module("io.daytona:api-client")).using(project(":"))
    }
}

includeBuild("../toolbox-api-client-java") {
    name = "toolbox-api-client"
    dependencySubstitution {
        substitute(module("io.daytona:toolbox-api-client")).using(project(":"))
    }
}
