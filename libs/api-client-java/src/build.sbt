lazy val root = (project in file(".")).
  settings(
    organization := "io.daytona",
    name := "daytona-api-client",
    version := "1.0",
    scalaVersion := "2.11.4",
    scalacOptions ++= Seq("-feature"),
    javacOptions in compile ++= Seq("-Xlint:deprecation"),
    publishArtifact in (Compile, packageDoc) := false,
    resolvers += Resolver.mavenLocal,
    libraryDependencies ++= Seq(
      "io.swagger" % "swagger-annotations" % "1.6.5",
      "com.squareup.okhttp3" % "okhttp" % "4.12.0",
      "com.squareup.okhttp3" % "logging-interceptor" % "4.12.0",
      "com.google.code.gson" % "gson" % "2.9.1",
      "org.apache.commons" % "commons-lang3" % "3.17.0",
      "jakarta.ws.rs" % "jakarta.ws.rs-api" % "2.1.6",
      "org.openapitools" % "jackson-databind-nullable" % "0.2.6",
      "io.gsonfire" % "gson-fire" % "1.9.0" % "compile",
      "jakarta.annotation" % "jakarta.annotation-api" % "1.3.5" % "compile",
      "com.google.code.findbugs" % "jsr305" % "3.0.2" % "compile",
      "jakarta.annotation" % "jakarta.annotation-api" % "1.3.5" % "compile",
      "org.junit.jupiter" % "junit-jupiter-api" % "5.10.3" % "test",
      "com.novocode" % "junit-interface" % "0.10" % "test",
      "org.mockito" % "mockito-core" % "3.12.4" % "test"
    )
  )
