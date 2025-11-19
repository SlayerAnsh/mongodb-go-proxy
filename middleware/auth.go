package middleware

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

// getAPISecret extracts the API secret from request headers
func getAPISecret(c echo.Context) string {
	// Get api-key from header

	apiKey := c.Request().Header.Get("api-key")
	return apiKey
}

// APISecretAuth validates the api-secret header (legacy, use ReadAuth or WriteAuth)
func APISecretAuth(expectedSecret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			apiKey := getAPISecret(c)

			if apiKey == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "api-key header is required",
				})
			}

			if apiKey != expectedSecret {
				return c.JSON(http.StatusForbidden, map[string]string{
					"error": "Invalid api-key",
				})
			}

			return next(c)
		}
	}
}

// ReadAuth validates the api-secret header for read operations
// Accepts both API_SECRET and READONLY_API_SECRET
func ReadAuth(apiSecret, readOnlyAPISecret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			providedSecret := getAPISecret(c)

			if providedSecret == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "api-key header is required",
				})
			}

			// Accept API_SECRET for read operations
			if providedSecret == apiSecret {
				return next(c)
			}

			// Also accept READONLY_API_SECRET if it's configured
			if readOnlyAPISecret != "" && providedSecret == readOnlyAPISecret {
				return next(c)
			}

			// Invalid secret
			return c.JSON(http.StatusForbidden, map[string]string{
				"error": "Invalid api-key",
			})
		}
	}
}

// WriteAuth validates the api-secret header for write operations
// Only accepts API_SECRET (not READONLY_API_SECRET)
func WriteAuth(apiSecret string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			providedSecret := getAPISecret(c)

			if providedSecret == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "api-key header is required",
				})
			}

			// Only accept API_SECRET for write operations
			if providedSecret != apiSecret {
				return c.JSON(http.StatusForbidden, map[string]string{
					"error": "Invalid api-key. Write operations require full API access.",
				})
			}

			return next(c)
		}
	}
}
