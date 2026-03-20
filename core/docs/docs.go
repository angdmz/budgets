// Package docs provides Swagger documentation configuration.
// All values are configurable via environment variables or programmatic setup.
package docs

import (
	"os"

	"github.com/swaggo/swag"
)

type SwaggerConfig struct {
	Host             string
	BasePath         string
	Version          string
	Title            string
	Description      string
	ContactName      string
	ContactEmail     string
	LicenseName      string
	LicenseURL       string
	Schemes          []string
}

func DefaultConfig() SwaggerConfig {
	return SwaggerConfig{
		Host:             getEnvOrDefault("SWAGGER_HOST", ""),
		BasePath:         getEnvOrDefault("SWAGGER_BASE_PATH", "/api/v1"),
		Version:          getEnvOrDefault("API_VERSION", "1.0.0"),
		Title:            getEnvOrDefault("API_TITLE", "Budget Management System API"),
		Description:      getEnvOrDefault("API_DESCRIPTION", "A complete Budget Management System with multi-user support, SSO authentication, and group-based data isolation."),
		ContactName:      getEnvOrDefault("API_CONTACT_NAME", ""),
		ContactEmail:     getEnvOrDefault("API_CONTACT_EMAIL", ""),
		LicenseName:      getEnvOrDefault("API_LICENSE_NAME", "MIT"),
		LicenseURL:       getEnvOrDefault("API_LICENSE_URL", "https://opensource.org/licenses/MIT"),
		Schemes:          []string{"http", "https"},
	}
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

var SwaggerInfo *swag.Spec

func InitSwagger(cfg SwaggerConfig) {
	SwaggerInfo = &swag.Spec{
		Version:          cfg.Version,
		Host:             cfg.Host,
		BasePath:         cfg.BasePath,
		Schemes:          cfg.Schemes,
		Title:            cfg.Title,
		Description:      cfg.Description,
		InfoInstanceName: "swagger",
		SwaggerTemplate:  generateTemplate(cfg),
	}
	swag.Register(SwaggerInfo.InstanceName(), SwaggerInfo)
}

func generateTemplate(cfg SwaggerConfig) string {
	return `{
    "swagger": "2.0",
    "info": {
        "title": "` + cfg.Title + `",
        "description": "` + cfg.Description + `",
        "version": "` + cfg.Version + `",
        "contact": {
            "name": "` + cfg.ContactName + `",
            "email": "` + cfg.ContactEmail + `"
        },
        "license": {
            "name": "` + cfg.LicenseName + `",
            "url": "` + cfg.LicenseURL + `"
        }
    },
    "host": "` + cfg.Host + `",
    "basePath": "` + cfg.BasePath + `",
    "schemes": ["http", "https"],
    "securityDefinitions": {
        "BearerAuth": {
            "type": "apiKey",
            "name": "Authorization",
            "in": "header",
            "description": "Type 'Bearer' followed by a space and JWT token."
        }
    },
    "paths": {},
    "definitions": {}
}`
}

func init() {
	InitSwagger(DefaultConfig())
}
