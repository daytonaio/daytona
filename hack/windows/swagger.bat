:: Make sure you have Node, Java 11 and Swagger installed and available in your PATH before running as well as the npm package @openapitools/openapi-generator-cli
cd pkg\server
swag fmt
swag init --parseDependency --parseInternal --parseDepth 1 -o api\docs -g api\server.go

SET GO_POST_PROCESS_FILE="C:\Go\bin\gofmt.exe -w"
SET GIT_USER_ID=daytonaio
SET GIT_REPO_ID=daytona

npx.cmd --yes @openapitools/openapi-generator-cli generate -i api/docs/swagger.json -g go --package-name=serverapiclient --additional-properties=isGoSubmodule=true -o ../serverapiclient

DEL /Q /S ..\serverapiclient\.openapi-generator\FILES