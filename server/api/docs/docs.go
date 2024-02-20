// Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "contact": {},
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/server/config": {
            "get": {
                "description": "Get the server configuration",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "server"
                ],
                "summary": "Get the server configuration",
                "operationId": "GetConfig",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/types.ServerConfig"
                        }
                    }
                }
            },
            "post": {
                "description": "Set the server configuration",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "server"
                ],
                "summary": "Set the server configuration",
                "operationId": "SetConfig",
                "parameters": [
                    {
                        "description": "Server configuration",
                        "name": "config",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/types.ServerConfig"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/types.ServerConfig"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "types.FRPSConfig": {
            "type": "object",
            "properties": {
                "domain": {
                    "type": "string"
                },
                "port": {
                    "type": "integer"
                },
                "protocol": {
                    "type": "string"
                }
            }
        },
        "types.ServerConfig": {
            "type": "object",
            "properties": {
                "apiPort": {
                    "type": "integer"
                },
                "frps": {
                    "$ref": "#/definitions/types.FRPSConfig"
                },
                "headscalePort": {
                    "type": "integer"
                },
                "id": {
                    "type": "string"
                },
                "pluginRegistryUrl": {
                    "type": "string"
                },
                "pluginsDir": {
                    "type": "string"
                },
                "serverDownloadUrl": {
                    "type": "string"
                }
            }
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "0.1.0",
	Host:             "localhost:3000",
	BasePath:         "/",
	Schemes:          []string{"http"},
	Title:            "Daytona Server API",
	Description:      "Daytona Server API",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
