// Package docs Code generated by swaggo/swag. DO NOT EDIT
package docs

import "github.com/swaggo/swag"

const docTemplate = `{
    "schemes": {{ marshal .Schemes }},
    "swagger": "2.0",
    "info": {
        "description": "{{escape .Description}}",
        "title": "{{.Title}}",
        "termsOfService": "http://swagger.io/terms/",
        "contact": {
            "name": "JffYang",
            "url": "https://github.com/kungze/wovenet/issues",
            "email": "jeffyang512@163.com"
        },
        "license": {
            "name": "MIT",
            "url": "https://github.com/kungze/wovenet/blob/main/LICENSE"
        },
        "version": "{{.Version}}"
    },
    "host": "{{.Host}}",
    "basePath": "{{.BasePath}}",
    "paths": {
        "/localExposedApps": {
            "get": {
                "description": "List all local exposed apps",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "localExposedApps"
                ],
                "summary": "List local exposed apps",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/app.LocalExposedAppModel"
                            }
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/restfulapi.HTTPError"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/restfulapi.HTTPError"
                        }
                    }
                }
            }
        },
        "/localExposedApps/{appName}": {
            "get": {
                "description": "Show a local exposed app by appName",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "localExposedApps"
                ],
                "summary": "Show a local exposed app",
                "parameters": [
                    {
                        "type": "string",
                        "description": "local exposed app name",
                        "name": "appName",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/app.LocalExposedAppModel"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/restfulapi.HTTPError"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/restfulapi.HTTPError"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/restfulapi.HTTPError"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/restfulapi.HTTPError"
                        }
                    }
                }
            }
        },
        "/remoteApps": {
            "get": {
                "description": "List all remote apps",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "remoteApps"
                ],
                "summary": "List remote apps",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/app.RemoteAppModel"
                            }
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/restfulapi.HTTPError"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/restfulapi.HTTPError"
                        }
                    }
                }
            }
        },
        "/remoteApps/{appName}": {
            "get": {
                "description": "Show a remote app by appName",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "remoteApps"
                ],
                "summary": "Show a remote app",
                "parameters": [
                    {
                        "type": "string",
                        "description": "remote app name",
                        "name": "appName",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/app.RemoteAppModel"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/restfulapi.HTTPError"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/restfulapi.HTTPError"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/restfulapi.HTTPError"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/restfulapi.HTTPError"
                        }
                    }
                }
            }
        },
        "/remoteSites": {
            "get": {
                "description": "Get all remote sites",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "remoteSites"
                ],
                "summary": "List remote sites",
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "type": "array",
                            "items": {
                                "$ref": "#/definitions/site.RemoteSiteModel"
                            }
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/restfulapi.HTTPError"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/restfulapi.HTTPError"
                        }
                    }
                }
            }
        },
        "/remoteSites/{siteName}": {
            "get": {
                "description": "Get a remote site's details by siteName",
                "consumes": [
                    "application/json"
                ],
                "produces": [
                    "application/json"
                ],
                "tags": [
                    "remoteSites"
                ],
                "summary": "Get a remote site's details",
                "parameters": [
                    {
                        "type": "string",
                        "description": "remote site name",
                        "name": "siteName",
                        "in": "path",
                        "required": true
                    }
                ],
                "responses": {
                    "200": {
                        "description": "OK",
                        "schema": {
                            "$ref": "#/definitions/site.RemoteSiteModel"
                        }
                    },
                    "400": {
                        "description": "Bad Request",
                        "schema": {
                            "$ref": "#/definitions/restfulapi.HTTPError"
                        }
                    },
                    "401": {
                        "description": "Unauthorized",
                        "schema": {
                            "$ref": "#/definitions/restfulapi.HTTPError"
                        }
                    },
                    "404": {
                        "description": "Not Found",
                        "schema": {
                            "$ref": "#/definitions/restfulapi.HTTPError"
                        }
                    },
                    "500": {
                        "description": "Internal Server Error",
                        "schema": {
                            "$ref": "#/definitions/restfulapi.HTTPError"
                        }
                    }
                }
            }
        }
    },
    "definitions": {
        "app.LocalExposedApp": {
            "type": "object",
            "properties": {
                "name": {
                    "type": "string"
                }
            }
        },
        "app.LocalExposedAppModel": {
            "type": "object",
            "properties": {
                "addressRange": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                },
                "appName": {
                    "type": "string"
                },
                "appSocket": {
                    "type": "string"
                },
                "mode": {
                    "type": "string"
                },
                "portRange": {
                    "type": "array",
                    "items": {
                        "type": "string"
                    }
                }
            }
        },
        "app.RemoteAppModel": {
            "type": "object",
            "properties": {
                "appName": {
                    "type": "string"
                },
                "appSocket": {
                    "type": "string"
                },
                "localSocket": {
                    "type": "string"
                },
                "siteName": {
                    "type": "string"
                }
            }
        },
        "restfulapi.HTTPError": {
            "type": "object",
            "properties": {
                "code": {
                    "type": "integer",
                    "example": 400
                },
                "message": {
                    "type": "string",
                    "example": "status bad request"
                }
            }
        },
        "site.RemoteSiteModel": {
            "type": "object",
            "properties": {
                "exposedApps": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/app.LocalExposedApp"
                    }
                },
                "siteName": {
                    "type": "string"
                },
                "tunnelSockets": {
                    "type": "array",
                    "items": {
                        "$ref": "#/definitions/tunnel.SocketInfo"
                    }
                }
            }
        },
        "tunnel.SocketInfo": {
            "type": "object",
            "properties": {
                "address": {
                    "type": "string"
                },
                "dynamicPublicAddress": {
                    "type": "boolean"
                },
                "id": {
                    "type": "string"
                },
                "port": {
                    "type": "integer"
                },
                "protocol": {
                    "$ref": "#/definitions/tunnel.TransportProtocol"
                }
            }
        },
        "tunnel.TransportProtocol": {
            "type": "string",
            "enum": [
                "quic",
                "SCTP"
            ],
            "x-enum-varnames": [
                "QUIC",
                "SCTP"
            ]
        }
    },
    "securityDefinitions": {
        "BasicAuth": {
            "type": "basic"
        }
    }
}`

// SwaggerInfo holds exported Swagger Info so clients can modify it
var SwaggerInfo = &swag.Spec{
	Version:          "1.0",
	Host:             "",
	BasePath:         "/api/v1",
	Schemes:          []string{},
	Title:            "Wovenet API",
	Description:      "This is wovenet api definitions",
	InfoInstanceName: "swagger",
	SwaggerTemplate:  docTemplate,
	LeftDelim:        "{{",
	RightDelim:       "}}",
}

func init() {
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}
