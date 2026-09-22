# LayoutMasterGO

<p align="center">
  <a href="https://golang.org/"><img src="https://img.shields.io/badge/Go-1.22+-00ADD8?style=for-the-badge&logo=go&logoColor=white" alt="Go Version" /></a>
  <a href="https://github.com/gin-gonic/gin"><img src="https://img.shields.io/badge/Gin_Framework-1.12-008ECF?style=for-the-badge&logo=go&logoColor=white" alt="Gin" /></a>
  <a href="https://gorm.io/"><img src="https://img.shields.io/badge/GORM-1.31-FF4154?style=for-the-badge&logo=go&logoColor=white" alt="GORM" /></a>
  <a href="https://github.com/swaggo/swag"><img src="https://img.shields.io/badge/Swagger-OpenAPI_2.0-85EA2D?style=for-the-badge&logo=swagger&logoColor=black" alt="Swagger" /></a>
  <a href="https://github.com/Frukas/LayoutMasterGO/commits/main"><img src="https://img.shields.io/github/last-commit/Frukas/LayoutMasterGO?style=for-the-badge&logo=git&logoColor=white&color=blue" alt="Last Commit" /></a>
  <a href="https://github.com/Frukas/LayoutMasterGO/blob/main/LICENSE"><img src="https://img.shields.io/badge/License-MIT-green?style=for-the-badge" alt="License" /></a>
</p>

<p align="center">
  <strong>High-performance Go backend and Alpine.js SPA for managing physical containers, products, and bulk Excel imports using Clean Architecture, Gin, GORM, and SQLite.</strong>
</p>

---

## 📌 Table of Contents

- [Overview](#-overview)
- [Key Features](#-key-features)
- [Architecture & Tech Stack](#-architecture--tech-stack)
- [Directory Structure](#-directory-structure)
- [Prerequisites](#-prerequisites)
- [Getting Started](#-getting-started)
  - [1. Clone the Repository](#1-clone-the-repository)
  - [2. Environment Configuration](#2-environment-configuration)
  - [3. Install Dependencies](#3-install-dependencies)
  - [4. Running the Web Server](#4-running-the-web-server)
  - [5. Running the Batch Excel Importer](#5-running-the-batch-excel-importer)
- [API Documentation & Endpoints](#-api-documentation--endpoints)
  - [Interactive Swagger UI](#interactive-swagger-ui)
  - [Container Endpoints (`/api/v1/containers`)](#container-endpoints-apiv1containers)
  - [Container Item Endpoints (`/api/v1/containers/:id/items`)](#container-item-endpoints-apiv1containersiditems)
  - [Product Endpoints (`/api/v1/products`)](#product-endpoints-apiv1products)
  - [Regenerating Swagger Docs](#regenerating-swagger-docs)
- [Single Page Application (SPA) Interface](#-single-page-application-spa-interface)
- [Excel Importer Architecture & Worker Pool](#-excel-importer-architecture--worker-pool)
- [Testing & Quality Assurance](#-testing--quality-assurance)
- [Contributing](#-contributing)
- [License](#-license)

---

## 🚀 Overview

**LayoutMasterGO** is a modular Go application designed to track and manage physical storage lots, shipping containers, and product inventory. Developed following **Clean Architecture** principles (Controllers, Services, Repositories, and Domain Models), it decouples HTTP routing, business logic, and database persistence.

The system features:
- A production **RESTful JSON API** built on [Gin](https://github.com/gin-gonic/gin) and [GORM](https://gorm.io/).
- An embedded **Alpine.js & Bootstrap 5 Single Page Application (SPA)** bundled directly into the compiled Go binary using `embed.FS`.
- A high-throughput, concurrent **Excel batch ingestion engine** using [Excelize](https://github.com/xuri/excelize) and a producer-worker goroutine pool for processing complex spreadsheet matrices into SQLite database records.

---

## ✨ Key Features

- **Container Management**:
  - Full CRUD operations for containers (`Name`, `Data` timestamp, `Status`).
  - Search and pagination support (`?search=...&page=1&pageSize=10`).
  - Item associations: link products with quantities inside containers through a relational pivot table (`ItemContainer`).

- **Product Catalog**:
  - Full CRUD operations for products with unique product names and standard barcode identifiers (`JanCode`).
  - Filter and search products with pagination.

- **Dual Execution Modes**:
  - **`cmd/server`**: Production HTTP API server + embedded web UI + Swagger documentation.
  - **`cmd/importer`**: Standalone command-line importer designed for headless, fast ingestion of matrix-style `.xlsx` spreadsheets.

- **Concurrent Excel Importer (Worker Pool)**:
  - **Phase 1 (Synchronous)**: Ingests all products, normalizes JAN codes, and builds an in-memory cache (`janToIDCache`).
  - **Phase 2 (Concurrent Worker Pool)**: Dispatches container column matrices to a pool of concurrent goroutine workers to batch-insert container and item records with SQLite WAL mode.

- **Embedded Web Client**:
  - Zero-build web dashboard powered by **Alpine.js** and **Bootstrap 5**, served directly from Go (`web/web.go`) with zero external frontend build step.

- **Automated Swagger / OpenAPI 2.0 Docs**:
  - Fully documented endpoints, request/response models, and error responses at `/swagger/index.html`.

---

## 🛠 Architecture & Tech Stack

| Layer / Component | Technology | Purpose |
|---|---|---|
| **Language** | [Go 1.22+](https://golang.org/) | Backend execution runtime |
| **HTTP Framework** | [Gin Web Framework](https://github.com/gin-gonic/gin) (v1.12) | High-performance routing, middleware & handlers |
| **ORM & Database** | [GORM](https://gorm.io/) (v1.31) + [glebarez/sqlite](https://github.com/glebarez/sqlite) | Database abstraction & pure-Go SQLite driver |
| **Spreadsheet Engine** | [Excelize v2](https://github.com/xuri/excelize) (v2.11) | Parsing and streaming `.xlsx` spreadsheets |
| **API Documentation** | [swaggo/swag](https://github.com/swaggo/swag) + [gin-swagger](https://github.com/swaggo/gin-swagger) | OpenAPI 2.0 specification and Swagger UI |
| **Web Frontend** | Alpine.js + Bootstrap 5 + Bootstrap Icons | Lightweight reactive SPA embedded via `embed.FS` |
| **Environment Config** | [joho/godotenv](https://github.com/joho/godotenv) | Environment variable management via `.env` |
| **Testing & Mocks** | `testing`, [stretchr/testify](https://github.com/stretchr/testify) | Unit, service, and controller testing with repository mocks |

---

## 📂 Directory Structure

```text
LayoutMasterGO/
└── LayoutMasterGo/
    ├── cmd/
    │   ├── server/              # Web API and static frontend entry point
    │   │   ├── main.go
    │   │   └── app.db           # SQLite database file (default)
    │   └── importer/            # Headless Excel batch importer entry point
    │       └── main.go
    ├── docs/                    # Generated Swagger & OpenAPI 2.0 specifications
    │   ├── docs.go
    │   ├── swagger.json
    │   └── swagger.yaml
    ├── internal/
    │   ├── api/                 # API routing & middlewares
    │   │   ├── router.go        # Route registration & static asset mapping
    │   │   └── middleware/      # CORS & request middleware
    │   ├── controller/          # HTTP handlers (Gin controllers)
    │   │   ├── container_controller.go
    │   │   ├── container_controller_test.go
    │   │   ├── product_controller.go
    │   │   └── product_controller_test.go
    │   ├── database/            # Database initialization and dialector setup
    │   │   └── database.go
    │   ├── importer/            # Excel parser and worker pool logic
    │   │   ├── importer.go      # Orchestrator (Run pipeline)
    │   │   ├── parser.go        # Spreadsheet row/column parsing & sanitization
    │   │   └── worker.go        # Goroutine worker pool & container dispatch
    │   ├── mocks/               # Mockery/test mocks for repositories
    │   ├── models/              # Domain entities (Product, Container, ItemContainer)
    │   │   └── models.go
    │   ├── repository/          # GORM data persistence layer
    │   │   ├── container_repo.go
    │   │   ├── item_container_repo.go
    │   │   ├── product_repository.go
    │   │   └── pagination.go
    │   └── service/             # Core business rules & validation logic
    │       ├── container_service.go
    │       └── product_service.go
    ├── web/                     # Embedded SPA web assets
    │   ├── index.html           # Main application shell
    │   ├── js/                  # Alpine.js application controllers
    │   │   ├── app.js           # API client, toast notifications & routing
    │   │   ├── containers.js    # Container management component
    │   │   └── products.js      # Product catalog component
    │   ├── templates/           # Sub-templates (containers.html, products.html)
    │   ├── img/                 # Logos and icons (favicon.png)
    │   └── web.go               # Go embed.FS filesystem bundle
    ├── go.mod                   # Go module definitions
    ├── go.sum                   # Dependency lockfile
    └── lembrete.txt             # Project notes & roadmap
```

---

## 📋 Prerequisites

- **Go**: Version `1.22.0` or higher ([Download Go](https://go.dev/dl/))
- **Git**: Version control
- (Optional) **Swag CLI**: For updating API documentation (`go install github.com/swaggo/swag/cmd/swag@latest`)

---

## 🏁 Getting Started

### 1. Clone the Repository

```bash
git clone https://github.com/Frukas/LayoutMasterGO.git
cd LayoutMasterGO/LayoutMasterGo
```

### 2. Environment Configuration

The application loads environment variables via `.env`. If not provided, it falls back to system defaults.

Create a `.env` file in the project root or in `cmd/server/`:

```env
PORT=8080
DB_PATH=app.db
```

### 3. Install Dependencies

Download and verify modules:

```bash
go mod tidy
go mod download
```

### 4. Running the Web Server

Start the primary HTTP server and REST API:

```bash
go run cmd/server/main.go
```

By default, the server binds to port **8080**:
- **Web SPA Interface**: [http://localhost:8080/](http://localhost:8080/)
- **Swagger Documentation**: [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)

### 5. Running the Batch Excel Importer

To ingest warehouse or inventory spreadsheets into the database:

```bash
go run cmd/importer/main.go path/to/data.xlsx
```

*(If no file parameter is passed, it defaults to looking for `data.xlsx` in the current directory).*

---

## 📖 API Documentation & Endpoints

### Interactive Swagger UI

When the server is running, explore and test endpoints interactively at:

```text
http://localhost:8080/swagger/index.html
```

---

### Container Endpoints (`/api/v1/containers`)

Base path: `/api/v1`

| HTTP Method | Endpoint | Description | Query / Path Parameters | Request Body |
|---|---|---|---|---|
| `POST` | `/api/v1/containers` | Create a new container | — | `{"name": "Lote A-100", "data": "2026-09-09T10:00:00Z", "status": "Active"}` |
| `GET` | `/api/v1/containers` | List containers with pagination and search | `page` (int, default `1`)<br>`pageSize` (int, default `10`)<br>`search` (string) | — |
| `GET` | `/api/v1/containers/:id` | Get container details by ID (including items) | `id` (uint, required) | — |
| `PUT` | `/api/v1/containers/:id` | Update container information | `id` (uint, required) | `{"name": "Lote A-200", "data": "2026-09-09T10:00:00Z", "status": "Finished"}` |
| `DELETE` | `/api/v1/containers/:id` | Delete a container by ID | `id` (uint, required) | — |

---

### Container Item Endpoints (`/api/v1/containers/:id/items`)

Manage product allocations inside specific containers:

| HTTP Method | Endpoint | Description | Path Parameters | Request Body |
|---|---|---|---|---|
| `POST` | `/api/v1/containers/:id/items` | Add a product to a container | `id` (uint, container ID) | `{"product_id": 1, "quantity": 10}` |
| `PUT` | `/api/v1/containers/:id/items/:product_id` | Update quantity of a product in container | `id` (uint, container ID)<br>`product_id` (uint, product ID) | `{"quantity": 25}` |
| `DELETE` | `/api/v1/containers/:id/items/:product_id` | Remove a product from a container | `id` (uint, container ID)<br>`product_id` (uint, product ID) | — |

---

### Product Endpoints (`/api/v1/products`)

| HTTP Method | Endpoint | Description | Query / Path Parameters | Request Body |
|---|---|---|---|---|
| `POST` | `/api/v1/products` | Create a new product | — | `{"name": "Detergente Líquido", "jan_code": "4901234567890"}` |
| `GET` | `/api/v1/products` | List products with pagination and search | `page` (int, default `1`)<br>`pageSize` (int, default `10`)<br>`search` (string) | — |
| `GET` | `/api/v1/products/:id` | Get product details by ID | `id` (uint, required) | — |
| `PUT` | `/api/v1/products/:id` | Update product information | `id` (uint, required) | `{"name": "Detergente Concentrado", "jan_code": "4901234567890"}` |
| `DELETE` | `/api/v1/products/:id` | Delete a product by ID | `id` (uint, required) | — |

---

### Static & Swagger Routes

| HTTP Method | Route | Description |
|---|---|---|
| `GET` | `/` | Serves the Single Page Application dashboard (`index.html`) |
| `GET` | `/static/*filepath` | Serves static assets, scripts, and HTML templates from `web/` |
| `GET` | `/swagger/*any` | Swagger UI documentation handler |

---

### Regenerating Swagger Docs

To re-generate API specifications after modifying annotations in `cmd/server/main.go` or `internal/controller/`:

```bash
swag init -g cmd/server/main.go -o docs
```

---

## 🖥 Single Page Application (SPA) Interface

The web interface is embedded directly into the Go executable via `embed.FS` (`web/web.go`), eliminating the need for Node.js, npm, or frontend bundling:

1. **Dashboard Shell (`web/index.html`)**: Navigation sidebar linking to **Containers** and **Products** views.
2. **Container Management (`web/templates/containers.html` & `web/js/containers.js`)**:
   - Paginated listing with real-time debounced search.
   - Container creation and editing modals.
   - Detail view inspecting container contents, adding items, updating quantities, and removing products.
3. **Product Management (`web/templates/products.html` & `web/js/products.js`)**:
   - Product registry with JAN barcode validation and search.
4. **Toast Notification System (`web/js/app.js`)**: Clean feedback on success or API error responses.

---

## 📊 Excel Importer Architecture & Worker Pool

The batch importer (`cmd/importer` and `internal/importer`) is tailored for matrix-style Excel layouts where rows list products (JAN codes) and columns represent containers:

```text
       Col A         Col B      ...    Col N       Col O       Col P
Row 6: JAN Code   | Product Name | ... | Cont. 1 | Cont. 2 | Cont. 3 ...
Row 7: 4901234567 | Product A    | ... | 50      | 12      | -
Row 8: 4901234568 | Product B    | ... | 100     | -       | 45
```

### Pipeline Flow:
1. **Spreadsheet Sanitization**: Trims whitespace, strips formatting commas (e.g. `1,200` &rarr; `1200`), and handles non-breaking spaces (`\u00A0`).
2. **Phase 1 — Product Synchronization**: Reads JAN codes and product names starting from Row 7, persists them via `ProductService`, and builds an in-memory lookup cache (`JAN -> ProductID`).
3. **Phase 2 — Producer-Consumer Worker Pool**:
   - Scans container columns starting at Column N (`colIndex = 13`).
   - Dispatches parsed `ContainerPayload` jobs through a buffered Go channel (`jobs`).
   - 3 concurrent goroutine workers consume jobs, create or resolve container records, and link items with their respective quantities.
   - Automatically detects end-of-matrix when consecutive empty headers are found.

---

## 🧪 Testing & Quality Assurance

LayoutMasterGO has extensive unit and integration tests across services, repositories, and controllers:

Run all tests:
```bash
go test -v ./...
```

Generate test coverage profile:
```bash
go test -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

HTML visual coverage report:
```bash
go tool cover -html=coverage.out -o coverage.html
```

---

## 🤝 Contributing

1. **Fork the Repository**
2. **Create a Feature Branch**:
   ```bash
   git checkout -b feature/new-feature
   ```
3. **Commit Your Changes**:
   ```bash
   git commit -m "feat: add container status transition rules"
   ```
4. **Push to the Branch**:
   ```bash
   git push origin feature/new-feature
   ```
5. **Open a Pull Request**

---

## 📄 License

This project is licensed under the **MIT License**. See the [LICENSE](LICENSE) file for details.
