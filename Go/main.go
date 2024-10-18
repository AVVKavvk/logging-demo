/*
This file shows example usage of log package
for logging as per platform logging guidelines (2023).
*/
package main

import (
	"context"
	"net/http"

	log "github.com/eencloud/goeen/log"
	uuid "github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

var logLevel = "DEBUG"

// EchoLoggingMiddleware logs request information and ensures request ID is set.
func EchoLoggingMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		requestID := c.Request().Header.Get("X-Request-ID")
		if requestID == "" {
			requestID = uuid.NewString()
			c.Request().Header.Set("X-Request-ID", requestID)
		}

		logger := log.DefaultV1Context.GetLogger("een.cloud.nexus", log.LevelInfo)
		logger.SetLevel(log.StringLevel(logLevel))

		// Set the request ID in the logger
		logger.SetRequestID(requestID)

		// Pass the logger in the context
		ctx := context.WithValue(c.Request().Context(), "logger", logger)
		c.SetRequest(c.Request().WithContext(ctx))

		return next(c)
	}
}

// GetLoggerFromContext retrieves the logger from the context.
func GetLoggerFromContext(ctx context.Context) *log.Logger {
	logger, ok := ctx.Value("logger").(*log.Logger)
	if !ok {
		return log.DefaultV1Context.GetLogger("een.cloud.nexus", log.LevelInfo)
	}
	return logger
}

func CustomFunction(ctx context.Context, data string) {
	logger := GetLoggerFromContext(ctx)

	logger.Infoxf(&log.XFields{"customKey": "customValue", "data": data}, "Custom function processing data")
}

func main() {
	e := echo.New()

	// Middleware setup
	e.Use(middleware.RequestID())
	e.Use(EchoLoggingMiddleware)

	// Main API route
	e.GET("/", func(c echo.Context) error {
		logger := GetLoggerFromContext(c.Request().Context())

		for i := 0; i < 3; i++ {
			uniqueId := uuid.New().String()

			_, err := http.Get("https://api.example.com/data")
			if err != nil {
				logger.Errorxf(&log.XFields{"key1": "Value1", "iter": i, "uniqueId": uniqueId, "err": err.Error()}, "API call failed: %v", err)
			}

			logger.Debugxf(&log.XFields{"key1": "Value1", "iter": i, "uniqueId": uniqueId}, "API response: %s")

			CustomFunction(c.Request().Context(), "customFunction")
		}

		return c.String(http.StatusOK, "API calls and custom function completed")
	})

	e.Start(":8081")
}
