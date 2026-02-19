package main

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/jcsoftdev/pulzifi-back/shared/logger"
	"go.uber.org/zap"
)

// @title Pulzifi API Documentation Hub
// @version 1.0
// @description Centralized API documentation portal for all Pulzifi microservices
// @host localhost
// @basePath /
// @schemes http https
func main() {
	logger.Info("Starting API Documentation Service")

	router := gin.Default()

	// ServiceDocs represents documentation for a service
	type ServiceDocs struct {
		Name string
		URL  string
		Port string
	}

	services := []ServiceDocs{
		{Name: "Organization", URL: "http://pulzifi-organization:8082", Port: "8082"},
		{Name: "Workspace", URL: "http://pulzifi-workspace:8083", Port: "8083"},
		{Name: "Page", URL: "http://pulzifi-page:8084", Port: "8084"},
		{Name: "Alert", URL: "http://pulzifi-alert:8085", Port: "8085"},
		{Name: "Monitoring", URL: "http://pulzifi-monitoring:8086", Port: "8086"},
		{Name: "Insight", URL: "http://pulzifi-insight:8087", Port: "8087"},
		{Name: "Report", URL: "http://pulzifi-report:8098", Port: "8098"},
		{Name: "Integration", URL: "http://pulzifi-integration:8089", Port: "8089"},
		{Name: "Usage", URL: "http://pulzifi-usage:8090", Port: "8090"},
		{Name: "Auth", URL: "http://pulzifi-auth:8081", Port: "8081"},
	}

	// CORS middleware
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Home page - list all services
	router.GET("/", func(c *gin.Context) {
		// Check service status
		type ServiceStatus struct {
			Name      string
			Available bool
			URL       string
		}

		serviceStatuses := []ServiceStatus{}
		for _, service := range services {
			resp, err := http.Get(service.URL + "/swagger/index.html")
			available := err == nil && resp.StatusCode == http.StatusOK
			if resp != nil && resp.Body != nil {
				resp.Body.Close()
			}
			serviceStatuses = append(serviceStatuses, ServiceStatus{
				Name:      service.Name,
				Available: available,
				URL:       service.Name,
			})
		}

		html := `
<!DOCTYPE html>
<html>
<head>
    <title>Pulzifi API Documentation</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            background: #f8f9fa;
            min-height: 100vh;
            display: flex;
            justify-content: center;
            align-items: center;
            padding: 20px;
        }
        .container {
            background: white;
            border-radius: 12px;
            box-shadow: 0 4px 20px rgba(0, 0, 0, 0.1);
            max-width: 900px;
            width: 100%;
            padding: 40px;
        }
        h1 {
            color: #333;
            margin-bottom: 10px;
            font-size: 2.5em;
        }
        .subtitle {
            color: #666;
            margin-bottom: 30px;
            font-size: 1.1em;
        }
        .legend {
            display: flex;
            gap: 20px;
            margin-bottom: 30px;
            padding-bottom: 20px;
            border-bottom: 1px solid #eee;
            font-size: 0.9em;
        }
        .legend-item {
            display: flex;
            align-items: center;
            gap: 8px;
        }
        .legend-dot {
            width: 12px;
            height: 12px;
            border-radius: 2px;
        }
        .legend-dot.available {
            background: #27ae60;
        }
        .legend-dot.unavailable {
            background: #bdc3c7;
        }
        .services-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(250px, 1fr));
            gap: 20px;
            margin-top: 30px;
        }
        .service-card {
            color: white;
            padding: 20px;
            border-radius: 8px;
            text-decoration: none;
            transition: all 0.3s ease;
            display: flex;
            flex-direction: column;
            justify-content: space-between;
            min-height: 150px;
            border-left: 4px solid rgba(0,0,0,0.1);
            position: relative;
        }
        .service-card.available {
            background: #ffffff;
            color: #333;
            box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
            border: 1px solid #e0e0e0;
            border-left: 4px solid #27ae60;
        }
        .service-card.available:hover {
            transform: translateY(-2px);
            box-shadow: 0 4px 16px rgba(0, 0, 0, 0.15);
        }
        .service-card.unavailable {
            background: #f0f0f0;
            color: #999;
            opacity: 1;
            cursor: not-allowed;
            border: 1px solid #e0e0e0;
            border-left: 4px solid #bdc3c7;
        }
        .service-card.unavailable:hover {
            transform: none;
        }
        .service-status {
            position: absolute;
            top: 10px;
            right: 10px;
            font-size: 1.2em;
        }
        .service-name {
            font-size: 1.3em;
            font-weight: 600;
            margin-bottom: 10px;
            color: #333;
        }
        .service-docs {
            font-size: 0.9em;
            opacity: 0.8;
            margin-bottom: 15px;
            color: #666;
        }
        .service-status-text {
            font-size: 0.8em;
            opacity: 0.7;
            margin-bottom: 10px;
            color: #888;
        }
        .docs-link {
            display: inline-block;
            background: #667eea;
            padding: 8px 16px;
            border-radius: 4px;
            transition: background 0.2s;
            text-decoration: none;
            color: white;
            border: none;
            font-size: 0.9em;
            font-weight: 500;
        }
        .service-card.available .docs-link {
            background: #667eea;
        }
        .service-card.available .docs-link:hover {
            background: #5568d3;
        }
        .service-card.unavailable .docs-link {
            background: #bdc3c7;
            cursor: not-allowed;
        }
        .footer {
            margin-top: 40px;
            padding-top: 20px;
            border-top: 1px solid #eee;
            text-align: center;
            color: #999;
            font-size: 0.9em;
        }
        .reload-hint {
            color: #f39c12;
            font-size: 0.85em;
            margin-top: 10px;
        }
    </style>
    <script>
        // Detect if we're behind a reverse proxy and adjust base path
        function getBasePath() {
            const path = window.location.pathname;
            if (path.includes('/api/docs')) {
                return '/api/docs';
            }
            return '';
        }
    </script>
</head>
<body>
    <div class="container">
        <h1>🚀 Pulzifi API Documentation</h1>
        <p class="subtitle">Explore API endpoints for all microservices</p>
        
        <div class="legend">
            <div class="legend-item">
                <div class="legend-dot available"></div>
                <span>Available</span>
            </div>
            <div class="legend-item">
                <div class="legend-dot unavailable"></div>
                <span>Building / Unavailable</span>
            </div>
        </div>
        
        <div class="services-grid" id="services">
`

		for _, status := range serviceStatuses {
			cardClass := "unavailable"
			statusEmoji := "⏳"
			statusText := "Building..."
			if status.Available {
				cardClass = "available"
				statusEmoji = "✅"
				statusText = "Available"
			}

			html += fmt.Sprintf(`
            <a href="javascript:void(0)" onclick="navigateToService('%s')" class="service-card %s" %s>
                <div>
                    <div class="service-status">%s</div>
                    <div class="service-name">%s Service</div>
                    <div class="service-docs">Browse API endpoints and try requests</div>
                    <div class="service-status-text">%s</div>
                </div>
                <span class="docs-link">View Documentation →</span>
            </a>
`, status.URL, cardClass, func() string {
				if !status.Available {
					return `onclick="event.stopPropagation()"`
				}
				return ""
			}(), statusEmoji, status.Name, statusText)
		}

		html += `
        </div>
        
        <div class="footer">
            <p>All services are documented with Swagger/OpenAPI 2.0</p>
            <p class="reload-hint">💡 Tip: Services are being built on first run. Refresh this page if you see "Building..." status</p>
        </div>
    </div>

    <script>
        function getBasePath() {
            const path = window.location.pathname;
            if (path.includes('/api/docs')) {
                return '/api/docs';
            }
            return '';
        }

        function navigateToService(serviceName) {
            const basePath = getBasePath();
            window.location.href = basePath + '/swagger/' + serviceName;
        }
    </script>
</body>
</html>
`
		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, html)
	})

	// Swagger UI proxy - redirect to specific service docs
	// Route to fetch and modify swagger-initializer.js
	router.GET("/swagger-initializer/:service", func(c *gin.Context) {
		serviceName := c.Param("service")

		var service *ServiceDocs
		for i := range services {
			if strings.EqualFold(services[i].Name, serviceName) {
				service = &services[i]
				break
			}
		}

		if service == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
			return
		}

		// Fetch the swagger-initializer.js from the service
		scriptURL := service.URL + "/swagger/swagger-initializer.js"
		resp, err := http.Get(scriptURL)
		if err != nil {
			logger.Error("Failed to fetch swagger-initializer", zap.String("service", serviceName), zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch swagger initializer"})
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			logger.Warn("Service returned non-200 for initializer", zap.String("service", serviceName), zap.Int("status", resp.StatusCode))
			c.JSON(http.StatusNotFound, gin.H{"error": "Swagger initializer not available"})
			return
		}

		// Read the script
		scriptBody, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.Error("Failed to read swagger-initializer", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read swagger initializer"})
			return
		}

		// Replace relative doc.json URL with absolute path
		scriptContent := string(scriptBody)
		scriptContent = strings.ReplaceAll(scriptContent, `url: "doc.json"`, `url: "/api/docs/swagger-assets/`+serviceName+`/doc.json"`)

		// Set proper headers and return the modified script
		c.Header("Content-Type", "application/javascript")
		c.String(http.StatusOK, scriptContent)
	})

	// Route to serve the swagger UI HTML for a specific service
	router.GET("/swagger/:service", func(c *gin.Context) {
		serviceName := c.Param("service")

		var service *ServiceDocs
		for i := range services {
			if strings.EqualFold(services[i].Name, serviceName) {
				service = &services[i]
				break
			}
		}

		if service == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Service not found"})
			return
		}

		// Fetch and serve the swagger UI from the service
		docsURL := service.URL + "/swagger/index.html"
		resp, err := http.Get(docsURL)
		if err != nil {
			logger.Error("Failed to fetch swagger docs", zap.String("service", serviceName), zap.Error(err))
			// Serve a friendly error page
			errorHTML := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>Documentation Unavailable</title>
    <style>
        body { font-family: Arial, sans-serif; display: flex; justify-content: center; align-items: center; height: 100vh; margin: 0; background: #f5f5f5; }
        .container { background: white; padding: 40px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); max-width: 500px; text-align: center; }
        h1 { color: #e74c3c; }
        p { color: #666; line-height: 1.6; }
        a { color: #667eea; text-decoration: none; }
        a:hover { text-decoration: underline; }
    </style>
</head>
<body>
    <div class="container">
        <h1>📚 Documentation Not Available</h1>
        <p><strong>%s Service</strong> documentation is not currently available.</p>
        <p>This could mean:</p>
        <ul style="text-align: left; display: inline-block;">
            <li>The service is still starting up</li>
            <li>The service hasn't been implemented yet</li>
            <li>The service is not responding</li>
        </ul>
        <p style="margin-top: 30px;"><a href="/">← Back to Hub</a></p>
    </div>
</body>
</html>
`, serviceName)
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.String(http.StatusOK, errorHTML)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			logger.Warn("Service returned non-200 status", zap.String("service", serviceName), zap.Int("status", resp.StatusCode))
			// Serve a friendly error page
			errorHTML := fmt.Sprintf(`
<!DOCTYPE html>
<html>
<head>
    <title>Documentation Not Ready</title>
    <style>
        body { font-family: Arial, sans-serif; display: flex; justify-content: center; align-items: center; height: 100vh; margin: 0; background: #f5f5f5; }
        .container { background: white; padding: 40px; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); max-width: 500px; text-align: center; }
        h1 { color: #f39c12; }
        p { color: #666; line-height: 1.6; }
        code { background: #f0f0f0; padding: 2px 6px; border-radius: 3px; font-family: monospace; }
        a { color: #667eea; text-decoration: none; }
        a:hover { text-decoration: underline; }
    </style>
</head>
<body>
    <div class="container">
        <h1>⚙️ Documentation Not Ready</h1>
        <p><strong>%s Service</strong> responded with status <code>%d</code>.</p>
        <p>The service may still be initializing. Try again in a few moments.</p>
        <p style="margin-top: 30px;"><a href="/">← Back to Hub</a></p>
    </div>
</body>
</html>
`, serviceName, resp.StatusCode)
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.String(http.StatusOK, errorHTML)
			return
		}

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			logger.Error("Failed to read swagger response", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read documentation"})
			return
		}

		// Modify the HTML to rewrite asset URLs correctly
		// The assets will be loaded via /api/docs/swagger-assets/:service/* proxy routes
		htmlContent := string(body)

		// Replace paths for CSS, JS, and JSON files
		// Handle various URL patterns in Swagger HTML/JS
		htmlContent = strings.ReplaceAll(htmlContent, `href="./`, `href="/api/docs/swagger-assets/`+serviceName+`/`)
		htmlContent = strings.ReplaceAll(htmlContent, `src="./`, `src="/api/docs/swagger-assets/`+serviceName+`/`)
		htmlContent = strings.ReplaceAll(htmlContent, `url("./`, `url("/api/docs/swagger-assets/`+serviceName+`/`)

		// Handle URL() references (CSS)
		htmlContent = strings.ReplaceAll(htmlContent, `url('./`, `url('/api/docs/swagger-assets/`+serviceName+`/`)

		// Handle data attributes and other paths
		htmlContent = strings.ReplaceAll(htmlContent, `data-url="./`, `data-url="/api/docs/swagger-assets/`+serviceName+`/`)

		// Replace the Swagger spec URL to point to swagger-assets instead of current path
		htmlContent = strings.ReplaceAll(htmlContent, `"./swagger.json"`, `"/api/docs/swagger-assets/`+serviceName+`/swagger.json"`)
		htmlContent = strings.ReplaceAll(htmlContent, `'./swagger.json'`, `'/api/docs/swagger-assets/`+serviceName+`/swagger.json'`)

		// Handle URLs without ./ prefix  (e.g., "index.css", "doc.json")
		htmlContent = strings.ReplaceAll(htmlContent, `href="index.css"`, `href="/api/docs/swagger-assets/`+serviceName+`/index.css"`)
		htmlContent = strings.ReplaceAll(htmlContent, `url("index`, `url("/api/docs/swagger-assets/`+serviceName+`/index`)
		htmlContent = strings.ReplaceAll(htmlContent, `url('index`, `url('/api/docs/swagger-assets/`+serviceName+`/index`)

		// Note: doc.json URL rewriting uses absolute paths set above
		// No additional rewriting needed for the relative URL

		c.Header("Content-Type", "text/html; charset=utf-8")
		c.String(http.StatusOK, htmlContent)
	})

	logger.Info("API Documentation Service running on :9000")
	if err := router.Run(":9000"); err != nil {
		logger.Error("Server error", zap.Error(err))
		panic(err)
	}
}
