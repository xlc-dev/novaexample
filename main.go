package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/xlc-dev/nova/nova"
)

// Item represents an individual item in our API with metadata about creation and status.
// This is the main data structure exposed through both JSON API and HTML interfaces.
type Item struct {
	ID        int       `json:"id" description:"Unique identifier for the item"`
	Name      string    `json:"name" minlength:"3" maxlength:"10" format:"alpha"`
	CreatedAt time.Time `json:"createdAt" description:"Timestamp when the item was created"`
	IsActive  bool      `json:"isActive,omitempty" description:"Indicates if the item is active"`
}

// NewItemInput represents the data structure clients send when creating new items.
// It contains only the fields that can be set during creation, excluding auto-generated fields.
type NewItemInput struct {
	Name     string `json:"name" description:"Name for the new item" minlength:"3" maxlength:"10" format:"alpha"`
	IsActive bool   `json:"isActive,omitempty" description:"Initial active status"`
}

// ErrorResponse provides a standardized JSON error response format.
// This ensures consistent error reporting across all API endpoints.
type ErrorResponse struct {
	Error string `json:"error" description:"Description of the error"`
}

// Global variables for simple in-memory storage
var (
	// items stores all items in memory using their ID as the key.
	items = make(map[int]Item)
	// nextItemID tracks the next available ID for new items.
	nextItemID = 0
	// mu protects concurrent access to the items map and nextItemID counter.
	mu sync.Mutex
)

//go:embed static/*
var staticFiles embed.FS

// getCommonStyles returns CSS styles used across multiple pages.
// This centralizes styling to maintain consistency across the application.
func getCommonStyles() string {
	return `
		:root {
			--primary-color: #f9a825;
			--primary-light: #ffcc66;
			--secondary-color: #87ceeb;
			--bg-color: #0a0f2a;
			--bg-gradient-end: #2a1a40;
			--text-color: #f0e6d2;
			--heading-color: #ffffff;
			--subtle-text: #cccccc;
			--subtle-bg: #101535;
			--border-color: #333858;
			--card-shadow: 0 4px 15px rgba(0, 0, 0, 0.4);
			--button-shadow: 0 2px 5px rgba(0, 0, 0, 0.3);
			--code-bg: #1e222c;
			--border-anim-speed: 2s;
		}

		*,
		*::before,
		*::after {
			box-sizing: border-box;
			margin: 0;
			padding: 0;
		}

		html {
			font-size: 16px;
			scroll-behavior: smooth;
			scroll-padding-top: 4rem;
		}

		body {
			font-family: "Inter", -apple-system, BlinkMacSystemFont, "Segoe UI",
				Roboto, "Helvetica Neue", Arial, sans-serif;
			line-height: 1.7;
			color: var(--text-color);
			background-color: var(--bg-color);
			background-image: linear-gradient(
				90deg,
				var(--bg-color) 0%,
				var(--bg-gradient-end) 100%
			);
			overflow-x: hidden;
		}

		.container {
			max-width: 1140px;
			width: 90%;
			margin: 0 auto;
			padding: 0 1rem;
		}

		.app-header {
			background-color: rgba(10, 15, 42, 0.8);
			backdrop-filter: blur(10px);
			border-bottom: 1px solid var(--border-color);
			padding: 1.5rem 0;
			margin-bottom: 2rem;
			text-align: center;
		}

		.app-header .logo {
			font-size: 1.8rem;
			font-weight: 700;
			color: var(--primary-light);
			text-decoration: none;
			display: inline-block;
		}

		.app-header .logo span {
			color: var(--secondary-color);
		}

		.main-nav ul {
			list-style: none;
			display: flex;
			justify-content: center;
			gap: 1.5rem;
			margin-top: 0.5rem;
		}

		.main-nav a {
			color: var(--text-color);
			text-decoration: none;
			font-weight: 500;
			transition: color 0.3s ease;
		}

		.main-nav a:hover, .main-nav a.active {
			color: var(--primary-color);
		}

		.btn {
			display: inline-block;
			padding: 0.8rem 1.8rem;
			border-radius: 50px;
			text-decoration: none;
			font-weight: 600;
			font-size: 1rem;
			transition: all 0.3s ease;
			cursor: pointer;
			border: none;
			box-shadow: var(--button-shadow);
			margin: 0.25rem;
		}

		.btn-primary {
			background-color: var(--primary-color);
			color: var(--bg-color);
		}

		.btn-primary:hover {
			background-color: var(--primary-light);
			transform: translateY(-2px);
			box-shadow: 0 4px 8px rgba(0, 0, 0, 0.3);
		}

		.btn-secondary {
			background-color: transparent;
			color: var(--primary-light);
			border: 1px solid var(--primary-light);
		}

		.btn-secondary:hover {
			background-color: rgba(249, 168, 37, 0.1);
			transform: translateY(-2px);
			box-shadow: 0 4px 8px rgba(0, 0, 0, 0.2);
		}

		.content-section {
			padding: 3rem 0;
			color: var(--heading-color);
		}

		.content-section .container {
			position: relative;
			z-index: 2;
		}

		.content-section h1, .content-section h2 {
			font-weight: 700;
			color: var(--heading-color);
			margin-bottom: 1.5rem;
			line-height: 1.3;
			text-align: center;
		}

		.content-section h1 {
			font-size: 2.8rem;
		}
		.content-section h2 {
			font-size: 2.2rem;
		}

		.content-section h1 span {
			color: var(--primary-color);
		}

		.content-section p, .content-section ul {
			font-size: 1.1rem;
			color: var(--text-color);
			max-width: 700px;
			margin: 0 auto 2rem auto;
			text-align: center;
		}

		.content-section ul {
			list-style-position: inside;
			padding-left: 0;
		}
		.content-section ul li {
			margin-bottom: 0.5rem;
		}

		.cta-buttons {
			display: flex;
			justify-content: center;
			gap: 1rem;
			margin-top: 2rem;
		}

		.table {
			width: 100%;
			border-collapse: collapse;
			margin: 2rem 0;
			background-color: var(--subtle-bg);
			border: 1px solid var(--border-color);
			border-radius: 8px;
			overflow: hidden;
			box-shadow: var(--card-shadow);
		}
		.table th, .table td {
			padding: 1rem;
			text-align: left;
			border-bottom: 1px solid var(--border-color);
			color: var(--text-color);
		}
		.table th {
			background-color: var(--bg-color);
			color: var(--primary-light);
			font-weight: 600;
		}
		.table tr:last-child td {
			border-bottom: none;
		}
		.table tr:hover td {
			background-color: var(--subtle-bg);
		}
		.table .btn {
			padding: 0.4rem 0.9rem;
			font-size: 0.9rem;
		}

		.form-group {
			margin-bottom: 1.5rem;
		}
		.form-group label {
			display: block;
			margin-bottom: 0.5rem;
			font-weight: 500;
			color: var(--primary-light);
		}
		.form-group input[type="text"],
		.form-group input[type="checkbox"],
		.form-group textarea {
			width: 100%;
			padding: 0.75rem;
			border: 1px solid var(--border-color);
			border-radius: 4px;
			box-sizing: border-box;
			background-color: var(--subtle-bg);
			color: var(--text-color);
			font-size: 1rem;
		}
		.form-group input[type="text"]:focus,
		.form-group textarea:focus {
			outline: none;
			border-color: var(--primary-color);
			box-shadow: 0 0 0 2px rgba(249, 168, 37, 0.3);
		}
		.form-group input[type="checkbox"] {
			width: auto;
			margin-right: 0.5rem;
			vertical-align: middle;
		}
		.form-actions {
			margin-top: 2rem;
			display: flex;
			gap: 1rem;
		}

		@media (max-width: 768px) {
			html {
				font-size: 15px;
			}
			.content-section h1 {
				font-size: 2.2rem;
			}
			.content-section h2 {
				font-size: 1.8rem;
			}
			.content-section p, .content-section ul {
				font-size: 1rem;
			}
			.cta-buttons {
				flex-direction: column;
				align-items: center;
			}
			.btn {
				width: 80%;
				max-width: 300px;
				text-align: center;
			}
			.form-actions {
				flex-direction: column;
			}
			.form-actions .btn {
				width: 100%;
			}
		}
	`
}

// setupRoutes configures all application routes including both HTML pages and JSON API endpoints.
// It demonstrates the dual nature of the Nova framework supporting both web pages and API responses.
func setupRoutes(router *nova.Router) {
	// Serve static files from the "static" directory
	staticFS, _ := fs.Sub(staticFiles, "static")
	router.Static("/static", staticFS)
	setupHTMLRoutes(router)
	setupAPIRoutes(router)
	setupDocumentationRoutes(router)
}

// setupHTMLRoutes configures routes that return HTML responses for web browser consumption.
// These routes demonstrate the HTML builder capabilities of the Nova framework.
func setupHTMLRoutes(router *nova.Router) {
	// Home page with navigation and feature overview
	router.GetFunc("/", handleHomePage, &nova.RouteOptions{
		Tags:        []string{"General"},
		Summary:     "Home page",
		Description: "Returns the main HTML page with navigation and feature overview.",
	})

	// Items list page showing all items in a table format
	router.GetFunc("/items", handleItemsListPage, &nova.RouteOptions{
		Tags:        []string{"General"},
		Summary:     "Items list page",
		Description: "Returns an HTML page showing all items in a table format.",
	})

	// Create item form page for adding new items
	router.GetFunc("/create", handleCreateItemPage, &nova.RouteOptions{
		Tags:        []string{"General"},
		Summary:     "Create item page",
		Description: "Returns an HTML form for creating new items.",
	})
}

// setupAPIRoutes configures JSON API endpoints with reduced boilerplate using enhanced handlers.
// These routes demonstrate clean JSON API development with automatic response handling.
func setupAPIRoutes(router *nova.Router) {
	api := router.Group("/api/v1")

	// List all items - simple JSON response with no boilerplate
	api.GetFunc("/items", handleGetItems, &nova.RouteOptions{
		Tags:        []string{"Items"},
		Summary:     "List all items",
		Description: "Retrieves a list of all items in the system.",
		Responses: map[int]nova.ResponseOption{
			http.StatusOK: {Description: "List of items", Body: []Item{}},
		},
	})

	// Get specific item by ID with automatic parameter extraction
	api.GetFunc("/items/{itemId}", handleGetItem, &nova.RouteOptions{
		Tags:        []string{"Items"},
		Summary:     "Get an item by ID",
		Description: "Retrieves details for a specific item using its unique identifier.",
		OperationID: "getItemById",
		Parameters: []nova.ParameterOption{{
			Name:        "itemId",
			In:          "path",
			Description: "The ID of the item to retrieve",
			Schema:      int(0),
		}},
		Responses: map[int]nova.ResponseOption{
			http.StatusOK:         {Description: "Item details", Body: &Item{}},
			http.StatusBadRequest: {Description: "Invalid item ID", Body: &ErrorResponse{}},
			http.StatusNotFound:   {Description: "Item not found", Body: &ErrorResponse{}},
		},
	})

	// Create new item with automatic binding and content negotiation
	api.PostFunc("/items", handleCreateItem, &nova.RouteOptions{
		Tags:        []string{"Items"},
		Summary:     "Create a new item",
		Description: "Adds a new item to the collection. Supports both JSON and form data input.",
		OperationID: "createItem",
		RequestBody: &NewItemInput{},
		Responses: map[int]nova.ResponseOption{
			http.StatusCreated:    {Description: "Item created successfully", Body: &Item{}},
			http.StatusBadRequest: {Description: "Invalid input", Body: &ErrorResponse{}},
		},
	})

	// Delete item by ID with automatic parameter extraction
	api.DeleteFunc("/items/{itemId}", handleDeleteItem, &nova.RouteOptions{
		Tags:        []string{"Items"},
		Summary:     "Delete an item",
		Description: "Removes an item from the collection using its unique identifier.",
		OperationID: "deleteItem",
		Parameters: []nova.ParameterOption{{
			Name:        "itemId",
			In:          "path",
			Description: "The ID of the item to delete",
			Schema:      int(0),
		}},
		Responses: map[int]nova.ResponseOption{
			http.StatusOK:         {Description: "Item deleted successfully"},
			http.StatusBadRequest: {Description: "Invalid item ID", Body: &ErrorResponse{}},
			http.StatusNotFound:   {Description: "Item not found", Body: &ErrorResponse{}},
		},
	})
}

// setupDocumentationRoutes configures routes for API documentation and error demonstration.
// These routes provide development and debugging utilities.
func setupDocumentationRoutes(router *nova.Router) {
	// Error demonstration endpoint - single line error response
	router.GetFunc("/error", func(rc *nova.ResponseContext) error {
		return rc.JSONError(500, "This is a demonstration error!")
	}, &nova.RouteOptions{
		Tags:        []string{"General"},
		Summary:     "Error demonstration",
		Description: "Demonstrates error handling and response formatting.",
	})

	// OpenAPI specification endpoint
	router.ServeOpenAPISpec("/openapi.json", nova.OpenAPIConfig{
		Title:       "Nova API with HTML & JSON",
		Version:     "1.0.0",
		Description: "This is an example API built with Nova that supports both HTML and JSON responses.",
	})

	// Swagger UI documentation interface
	router.ServeSwaggerUI("/docs")
}

// handleHomePage renders the main application homepage with navigation and feature overview.
// It demonstrates the HTML builder pattern for creating complete web pages.
func handleHomePage(rc *nova.ResponseContext) error {
	doc := nova.Document(
		nova.DocumentConfig{
			Title: "Nova App",
			HeadExtras: []nova.HTMLElement{
				nova.Favicon("/static/favicon.png"),
				nova.StyleTag(getCommonStyles()),
			},
		},
		nova.Header(
			nova.A("/", nova.Text("Nova"), nova.Span(nova.Text("App"))).Class("logo"),
		).Class("app-header"),
		nova.Main(
			nova.Section(
				nova.Div(
					nova.H1(nova.Text("Welcome to "), nova.Span(nova.Text("Nova"))),
					nova.P().Text("This application demonstrates key features of the Nova framework with a JSON API and server-rendered HTML pages."),
					nova.H2().Text("Explore Features"),
					nova.Div(
						nova.Link("/items", "View All Items").Class("btn btn-secondary"),
						nova.Link("/create", "Create New Item").Class("btn btn-secondary"),
						nova.Link("/api/v1/items", "Items JSON API").Class("btn btn-secondary"),
						nova.Link("/docs", "API Docs").Class("btn btn-secondary"),
					).Class("cta-buttons"),
				).Class("container"),
			).Class("content-section"),
		).Class("container"),
	)

	return rc.HTML(http.StatusOK, doc)
}

// handleItemsListPage renders a table view of all items with action buttons.
// It demonstrates dynamic HTML generation based on application data.
func handleItemsListPage(rc *nova.ResponseContext) error {
	// Get all items with thread safety
	mu.Lock()
	itemsList := make([]Item, 0, len(items))
	for _, item := range items {
		itemsList = append(itemsList, item)
	}
	mu.Unlock()

	// Build table rows dynamically
	rows := make([]nova.HTMLElement, 0, len(itemsList))
	for _, item := range itemsList {
		statusBadge := "Inactive"
		if item.IsActive {
			statusBadge = "Active"
		}
		row := nova.Tr(
			nova.Td().Text(strconv.Itoa(item.ID)),
			nova.Td().Text(item.Name),
			nova.Td().Text(item.CreatedAt.Format("Jan 02, 2006 15:04")),
			nova.Td().Text(statusBadge),
			nova.Td(
				nova.Link(fmt.Sprintf("/api/v1/items/%d", item.ID), "View JSON").
					Class("btn btn-secondary").
					Style("font-size: 0.8em; padding: 0.3em 0.6em;"),
			),
		)
		rows = append(rows, row)
	}

	// Create table or empty state message
	var tableContent nova.HTMLElement
	if len(rows) == 0 {
		tableContent = nova.P().Text("No items found. Create your first item to get started!")
	} else {
		tableContent = nova.Table(
			nova.Thead(
				nova.Tr(
					nova.Th().Text("ID"),
					nova.Th().Text("Name"),
					nova.Th().Text("Created At"),
					nova.Th().Text("Status"),
					nova.Th().Text("Actions"),
				),
			),
			nova.Tbody(rows...),
		).Class("table")
	}

	doc := nova.Document(
		nova.DocumentConfig{
			Title: "Items List",
			HeadExtras: []nova.HTMLElement{
				nova.Favicon("/static/favicon.png"),
				nova.StyleTag(getCommonStyles()),
			},
		},
		nova.Header(
			nova.A("/", nova.Text("Nova"), nova.Span(nova.Text("App"))).Class("logo"),
		).Class("app-header"),
		nova.Main(
			nova.Section(
				nova.Div(
					nova.H1().Text("Items Management"),
					tableContent,
					nova.Div(
						nova.Link("/", "Back to Home").Class("btn btn-secondary"),
						nova.Link("/create", "Create New Item").Class("btn btn-primary"),
					).Class("cta-buttons").Style("margin-top: 1rem; justify-content: center;"),
				).Class("container"),
			).Class("content-section"),
		).Class("container"),
	)

	return rc.HTML(http.StatusOK, doc)
}

// renderCreateItemForm renders the Create Item page,
// optionally showing an error message and re-populating fields.
func renderCreateItemForm(
	rc *nova.ResponseContext,
	input NewItemInput,
	errorMsg string,
) error {
	// Build the list of children for the container div
	children := []nova.HTMLElement{
		nova.H1().Text("Create New Item"),
	}

	// Only append an error banner if there is an error
	if errorMsg != "" {
		children = append(children,
			nova.Div(nova.Text(errorMsg)).
				Class("error-message").
				Style("color:#ff6b6b; font-weight:500; margin-bottom:1rem;"),
		)
	}

	// Name input, pre‐filled if needed
	nameInput := nova.TextInput("name").
		ID("name").
		Attr("required", "true").
		Attr("maxlength", "50").
		Attr("placeholder", "Enter item name")
	if input.Name != "" {
		nameInput.Attr("value", input.Name)
	}

	// Checkbox input, preserved on error
	checkbox := nova.CheckboxInput("isActive").ID("isActive")
	if input.IsActive {
		checkbox.Attr("checked", "true")
	}

	// Append the actual form to children
	children = append(children,
		nova.Form(
			nova.Div(
				nova.Label().Text("Name:").Attr("for", "name"),
				nameInput,
			).Class("form-group"),
			nova.Div(
				nova.Label(checkbox, nova.Text(" Item is active")),
			).Class("form-group"),
			nova.Div(
				nova.SubmitButton("Create Item").Class("btn btn-primary"),
				nova.A("/items", nova.Text("Cancel")).Class("btn btn-secondary"),
			).Class("form-actions"),
		).
			Attr("method", "POST").
			Attr("action", "/api/v1/items").
			Attr("enctype", "application/x-www-form-urlencoded"),
		nova.Br(),
		nova.Link("/", "Back to Home").
			Class("btn btn-secondary").
			Style("margin-top:1rem;"),
	)

	doc := nova.Document(
		nova.DocumentConfig{
			Title: "Create Item",
			HeadExtras: []nova.HTMLElement{
				nova.Favicon("/static/favicon.png"),
				nova.StyleTag(getCommonStyles()),
			},
		},
		nova.Header(
			// Corrected line:
			nova.A("/", nova.Text("Nova"), nova.Span(nova.Text("App"))).Class("logo"),
		).Class("app-header"),
		nova.Main(
			nova.Section(
				nova.Div(children...).Class("container"),
			).Class("content-section"),
		).Class("container"),
	)
	return rc.HTML(http.StatusOK, doc)
}

// handleCreateItemPage now simply calls our renderer with no error.
func handleCreateItemPage(rc *nova.ResponseContext) error {
	return renderCreateItemForm(rc, NewItemInput{}, "")
}

// handleGetItems returns a JSON list of all items - minimal boilerplate.
func handleGetItems(rc *nova.ResponseContext) error {
	mu.Lock()
	itemsList := make([]Item, 0, len(items))
	for _, item := range items {
		itemsList = append(itemsList, item)
	}
	mu.Unlock()

	return rc.JSON(http.StatusOK, itemsList)
}

// handleGetItem returns a specific item by ID with automatic parameter extraction.
func handleGetItem(rc *nova.ResponseContext) error {
	idStr := rc.URLParam("itemId")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return rc.JSONError(http.StatusBadRequest, "Invalid item ID format")
	}

	mu.Lock()
	item, exists := items[id]
	mu.Unlock()

	if !exists {
		return rc.JSONError(http.StatusNotFound, fmt.Sprintf("Item %d not found", id))
	}

	return rc.JSON(http.StatusOK, item)
}

// handleCreateItem binds & validates, then either returns JSON or re-renders the form.
func handleCreateItem(rc *nova.ResponseContext) error {
	var input NewItemInput
	if err := rc.BindValidated(&input); err != nil {
		// JSON clients get a JSON error
		if rc.WantsJSON() {
			return rc.JSONError(http.StatusBadRequest, "Invalid input: "+err.Error())
		}
		// HTML form clients see the form again with errors & previous data
		return renderCreateItemForm(rc, input, err.Error())
	}

	mu.Lock()
	nextItemID++
	id := nextItemID
	item := Item{
		ID:        id,
		Name:      input.Name,
		CreatedAt: time.Now().UTC(),
		IsActive:  input.IsActive,
	}
	items[id] = item
	mu.Unlock()

	if rc.WantsJSON() {
		return rc.JSON(http.StatusCreated, item)
	}
	return rc.Redirect(http.StatusFound, "/items")
}

// handleDeleteItem removes an item with automatic parameter extraction and error handling.
func handleDeleteItem(rc *nova.ResponseContext) error {
	idStr := rc.URLParam("itemId")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return rc.JSONError(http.StatusBadRequest, "Invalid item ID format")
	}

	mu.Lock()
	_, exists := items[id]
	if exists {
		delete(items, id)
	}
	mu.Unlock()

	if !exists {
		return rc.JSONError(http.StatusNotFound, fmt.Sprintf("Item %d not found", id))
	}

	return rc.JSON(http.StatusOK, map[string]string{
		"message": "Item deleted successfully",
		"id":      strconv.Itoa(id),
	})
}

// main demonstrates the complete Nova application setup with minimal configuration.
func main() {
	cli, err := nova.NewCLI(&nova.CLI{
		Name:        "nova-api",
		Version:     "1.0.0",
		Description: "Nova framework demonstration with HTML and JSON APIs, auto-reload, and minimal boilerplate",
		GlobalFlags: []nova.Flag{
			&nova.StringFlag{
				Name:    "host",
				Aliases: []string{"H"},
				Default: "localhost",
				Usage:   "Host to bind the server to",
			},
			&nova.IntFlag{
				Name:    "port",
				Aliases: []string{"p"},
				Default: 8080,
				Usage:   "Port to listen on",
			},
			&nova.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"V"},
				Usage:   "Enable verbose logging output",
			},
			&nova.BoolFlag{
				Name:    "watch",
				Aliases: []string{"w"},
				Usage:   "Enable live file watching and auto-reload during development",
			},
			&nova.StringFlag{
				Name:    "extensions",
				Aliases: []string{"e"},
				Default: ".go",
				Usage:   "File extensions to watch for changes (comma-separated)",
			},
			&nova.StringFlag{
				Name:    "log_format",
				Aliases: []string{"f"},
				Default: "text",
				Usage:   "Log output format (text or json)",
			},
			&nova.StringFlag{
				Name:    "log_level",
				Aliases: []string{"l"},
				Default: "info",
				Usage:   "Log level (debug, info, warn, error)",
			},
		},
		Action: func(ctx *nova.Context) error {
			// Initialize router
			router := nova.NewRouter()

			// Apply middleware stack
			router.Use(
				nova.RecoveryMiddleware(&nova.RecoveryConfig{}),
				nova.RequestIDMiddleware(&nova.RequestIDConfig{
					HeaderName: "X-Request-ID",
				}),
				nova.LoggingMiddleware(&nova.LoggingConfig{
					Logger:       log.Default(),
					LogRequestID: true,
				}),
				nova.SecurityHeadersMiddleware(nova.SecurityHeadersConfig{
					ContentTypeOptions:    "nosniff",
					FrameOptions:          "DENY",
					ReferrerPolicy:        "strict-origin-when-cross-origin",
					HSTSMaxAgeSeconds:     31536000, // 1 year
					HSTSIncludeSubdomains: func(b bool) *bool { return &b }(true),
					HSTSPreload:           false,
				}),
				nova.CORSMiddleware(nova.CORSConfig{
					AllowedOrigins:   []string{"*"},
					AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
					AllowedHeaders:   []string{"Content-Type", "Authorization", "X-Request-ID"},
					AllowCredentials: false,
					MaxAgeSeconds:    86400, // 24 hours
				}),
				nova.TrailingSlashRedirectMiddleware(nova.TrailingSlashRedirectConfig{
					AddSlash:     false,
					RedirectCode: http.StatusMovedPermanently,
				}),
			)

			// Setup all routes
			setupRoutes(router)

			// Start the Nova server
			return nova.Serve(ctx, router)
		},
	})

	if err != nil {
		log.Fatalf("Failed to initialize CLI: %v", err)
	}

	if err := cli.Run(os.Args); err != nil {
		log.Fatalf("Application error: %v", err)
	}
}
