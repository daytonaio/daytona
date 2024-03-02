// Package docs Code generated by swaggo/swag. DO NOT EDIT
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
        "/provider": {
            "get": {
                "description": "List providers",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "provider"
                ],
                "summary": "List providers",
                "operationId": "ListProviders",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/Provider"
                            }
                        }
                    }
                }
            }
        },
        "/provider/install": {
            "post": {
                "description": "Install a provider",
                "consumes": [
                    "application/json"
                ],
                "tags": [
                    "provider"
                ],
                "summary": "Install a provider",
                "operationId": "InstallProvider",
                "parameters": [
                    {
                        "description": "Provider to install",
                        "name": "provider",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/InstallProviderRequest"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK"
                    }
                }
            }
        },
        "/provider/{provider}/target-manifest": {
            "get": {
                "description": "Get provider target manifest",
                "tags": [
                    "provider"
                ],
                "summary": "Get provider target manifest",
                "operationId": "GetTargetManifest",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Provider name",
                        "name": "provider",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/ProviderTargetManifest"
                        }
                    }
                }
            }
        },
        "/provider/{provider}/uninstall": {
            "post": {
                "description": "Uninstall a provider",
                "consumes": [
                    "application/json"
                ],
                "tags": [
                    "provider"
                ],
                "summary": "Uninstall a provider",
                "operationId": "UninstallProvider",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Provider to uninstall",
                        "name": "provider",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK"
                    }
                }
            }
        },
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
                            "$ref": "#/definitions/ServerConfig"
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
                            "$ref": "#/definitions/ServerConfig"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/ServerConfig"
                        }
                    }
                }
            }
        },
        "/server/get-git-context/{gitUrl}": {
            "get": {
                "description": "Get Git context",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "server"
                ],
                "summary": "Get Git context",
                "operationId": "GetGitContext",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Git URL",
                        "name": "gitUrl",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/Repository"
                        }
                    }
                }
            }
        },
        "/server/network-key": {
            "post": {
                "description": "Generate a new authentication key",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "server"
                ],
                "summary": "Generate a new authentication key",
                "operationId": "GenerateNetworkKey",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/NetworkKey"
                        }
                    }
                }
            }
        },
        "/target": {
            "get": {
                "description": "List targets",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "target"
                ],
                "summary": "List targets",
                "operationId": "ListTargets",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/ProviderTarget"
                            }
                        }
                    }
                }
            },
            "put": {
                "description": "Set a target",
                "tags": [
                    "target"
                ],
                "summary": "Set a target",
                "operationId": "SetTarget",
                "parameters": [
                    {
                        "description": "Target to set",
                        "name": "target",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/ProviderTarget"
                        }
                    }
                ],
                "responses": {
                    "201": {
                        "description": "Created"
                    }
                }
            }
        },
        "/target/{target}": {
            "delete": {
                "description": "Remove a target",
                "tags": [
                    "target"
                ],
                "summary": "Remove a target",
                "operationId": "RemoveTarget",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Target name",
                        "name": "target",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "204": {
                        "description": "No Content"
                    }
                }
            }
        },
        "/workspace": {
            "get": {
                "description": "List workspaces",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "workspace"
                ],
                "summary": "List workspaces",
                "operationId": "ListWorkspaces",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/Workspace"
                            }
                        }
                    }
                }
            },
            "post": {
                "description": "Create a workspace",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "workspace"
                ],
                "summary": "Create a workspace",
                "operationId": "CreateWorkspace",
                "parameters": [
                    {
                        "description": "Create workspace",
                        "name": "workspace",
                        "in": "body",
                        "required": true,
                        "schema": {
                            "$ref": "#/definitions/CreateWorkspace"
                        }
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/Workspace"
                        }
                    }
                }
            }
        },
        "/workspace/{workspaceId}": {
            "get": {
                "description": "Get workspace info",
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "workspace"
                ],
                "summary": "Get workspace info",
                "operationId": "GetWorkspace",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Workspace ID",
                        "name": "workspaceId",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/Workspace"
                        }
                    }
                }
            },
            "delete": {
                "description": "Remove workspace",
                "tags": [
                    "workspace"
                ],
                "summary": "Remove workspace",
                "operationId": "RemoveWorkspace",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Workspace ID",
                        "name": "workspaceId",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK"
                    }
                }
            }
        },
        "/workspace/{workspaceId}/start": {
            "post": {
                "description": "Start workspace",
                "tags": [
                    "workspace"
                ],
                "summary": "Start workspace",
                "operationId": "StartWorkspace",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Workspace ID",
                        "name": "workspaceId",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK"
                    }
                }
            }
        },
        "/workspace/{workspaceId}/stop": {
            "post": {
                "description": "Stop workspace",
                "tags": [
                    "workspace"
                ],
                "summary": "Stop workspace",
                "operationId": "StopWorkspace",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Workspace ID",
                        "name": "workspaceId",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK"
                    }
                }
            }
        },
        "/workspace/{workspaceId}/{projectId}/start": {
            "post": {
                "description": "Start project",
                "tags": [
                    "workspace"
                ],
                "summary": "Start project",
                "operationId": "StartProject",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Workspace ID",
                        "name": "workspaceId",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Project ID",
                        "name": "projectId",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK"
                    }
                }
            }
        },
        "/workspace/{workspaceId}/{projectId}/stop": {
            "post": {
                "description": "Stop project",
                "tags": [
                    "workspace"
                ],
                "summary": "Stop project",
                "operationId": "StopProject",
                "parameters": [
                    {
                        "type": "string",
                        "description": "Workspace ID",
                        "name": "workspaceId",
                        "in": "path",
                        "required": true
                    },
                    {
                        "type": "string",
                        "description": "Project ID",
                        "name": "projectId",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK"
                    }
                }
            }
        }
    },
    "definitions": {
        "CreateWorkspace": {
            "type": "object",
            "properties": {
                "name": {
                    "type": "string"
                },
                "repositories": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/Repository"
                    }
                },
                "target": {
                    "type": "string"
                }
            }
        },
        "FRPSConfig": {
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
        "GitProvider": {
            "type": "object",
            "properties": {
                "id": {
                    "type": "string"
                },
                "token": {
                    "type": "string"
                },
                "username": {
                    "type": "string"
                }
            }
        },
        "InstallProviderRequest": {
            "type": "object",
            "properties": {
                "downloadUrls": {
                    "type": "object",
                    "additionalProperties": {
                        "type": "string"
                    }
                },
                "name": {
                    "type": "string"
                }
            }
        },
        "NetworkKey": {
            "type": "object",
            "properties": {
                "key": {
                    "type": "string"
                }
            }
        },
        "Project": {
            "type": "object",
            "properties": {
                "info": {
                    "$ref": "#/definitions/ProjectInfo"
                },
                "name": {
                    "type": "string"
                },
                "repository": {
                    "$ref": "#/definitions/Repository"
                },
                "target": {
                    "type": "string"
                },
                "workspaceId": {
                    "type": "string"
                }
            }
        },
        "ProjectInfo": {
            "type": "object",
            "properties": {
                "created": {
                    "type": "string"
                },
                "finished": {
                    "type": "string"
                },
                "isRunning": {
                    "type": "boolean"
                },
                "name": {
                    "type": "string"
                },
                "providerMetadata": {
                    "type": "string"
                },
                "started": {
                    "type": "string"
                },
                "workspaceId": {
                    "type": "string"
                }
            }
        },
        "Provider": {
            "type": "object",
            "properties": {
                "name": {
                    "type": "string"
                },
                "version": {
                    "type": "string"
                }
            }
        },
        "ProviderTarget": {
            "type": "object",
            "properties": {
                "name": {
                    "type": "string"
                },
                "options": {
                    "description": "JSON encoded map of options",
                    "type": "string"
                },
                "providerInfo": {
                    "$ref": "#/definitions/provider.ProviderInfo"
                }
            }
        },
        "ProviderTargetManifest": {
            "type": "object",
            "additionalProperties": {
                "$ref": "#/definitions/provider.ProviderTargetProperty"
            }
        },
        "Repository": {
            "type": "object",
            "properties": {
                "branch": {
                    "type": "string"
                },
                "owner": {
                    "type": "string"
                },
                "path": {
                    "type": "string"
                },
                "prNumber": {
                    "type": "integer"
                },
                "sha": {
                    "type": "string"
                },
                "source": {
                    "type": "string"
                },
                "url": {
                    "type": "string"
                }
            }
        },
        "ServerConfig": {
            "type": "object",
            "properties": {
                "apiPort": {
                    "type": "integer"
                },
                "frps": {
                    "$ref": "#/definitions/FRPSConfig"
                },
                "gitProviders": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/GitProvider"
                    }
                },
                "headscalePort": {
                    "type": "integer"
                },
                "id": {
                    "type": "string"
                },
                "providersDir": {
                    "type": "string"
                },
                "registryUrl": {
                    "type": "string"
                },
                "serverDownloadUrl": {
                    "type": "string"
                },
                "targetsFilePath": {
                    "type": "string"
                }
            }
        },
        "Workspace": {
            "type": "object",
            "properties": {
                "id": {
                    "type": "string"
                },
                "info": {
                    "$ref": "#/definitions/WorkspaceInfo"
                },
                "name": {
                    "type": "string"
                },
                "projects": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/Project"
                    }
                },
                "target": {
                    "type": "string"
                }
            }
        },
        "WorkspaceInfo": {
            "type": "object",
            "properties": {
                "name": {
                    "type": "string"
                },
                "projects": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/ProjectInfo"
                    }
                },
                "providerMetadata": {
                    "type": "string"
                }
            }
        },
        "provider.ProviderInfo": {
            "type": "object",
            "properties": {
                "name": {
                    "type": "string"
                },
                "version": {
                    "type": "string"
                }
            }
        },
        "provider.ProviderTargetProperty": {
            "type": "object",
            "properties": {
                "defaultValue": {
                    "description": "DefaultValue is converted into the appropriate type based on the Type\nIf the property is a FilePath, the DefaultValue is a path to a directory",
                    "type": "string"
                },
                "disabledPredicate": {
                    "description": "A regex string matched with the name of the target to determine if the property should be disabled\nIf the regex matches the target name, the property will be disabled\nE.g. \"^local$\" will disable the property for the local target",
                    "type": "string"
                },
                "inputMasked": {
                    "type": "boolean"
                },
                "options": {
                    "description": "Options is only used if the Type is ProviderTargetPropertyTypeOption",
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "type": {
                    "$ref": "#/definitions/provider.ProviderTargetPropertyType"
                }
            }
        },
        "provider.ProviderTargetPropertyType": {
            "type": "string",
            "enum": [
                "string",
                "option",
                "boolean",
                "int",
                "float",
                "file-path"
            ],
            "x-enum-varnames": [
                "ProviderTargetPropertyTypeString",
                "ProviderTargetPropertyTypeOption",
                "ProviderTargetPropertyTypeBoolean",
                "ProviderTargetPropertyTypeInt",
                "ProviderTargetPropertyTypeFloat",
                "ProviderTargetPropertyTypeFilePath"
            ]
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
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
