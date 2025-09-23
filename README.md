# ImageProcessor Service

A scalable image processing service with background task queue using Apache Kafka.
Users can upload images, which are processed asynchronously (resize, thumbnail generation, watermarking) and stored in a file storage.
A simple web interface allows uploading, monitoring processing status, and retrieving or deleting images.

---

## Features

* **HTTP API**

    * `POST /api/upload` — Upload an image for processing.
    * `GET /api/image/:id` — Retrieve the processed image by ID.
    * `GET /api/image/:id/meta` — Get image metadata by ID (status, filename, etc.).
    * `DELETE /api/image/:id` — Delete an image by ID.

* **Background image processing**

    * Resize
    * Generate thumbnails
    * Add watermarks

* **File storage**

    * Stores original and processed images separately.

* **Frontend**

    * Upload images via `<input type="file">`.
    * Display image processing status.
    * Show processed images.
    * Delete images.

---

## Directory Structure

```
.
├── backend/           # Backend service
│   ├── cmd/           # Entry points
│   ├── config/        # Configuration files
│   ├── internal/      # Internal application packages
│   │   ├── api/       # HTTP handlers, routers, server
│   │   │   assets/    # Fonts or static files
│   │   ├── config/    # Config parsing logic
│   │   ├── infra/     # Infrastructure (Kafka consumer/producer)
│   │   ├── kafka/     # Kafka message handlers
│   │   ├── middleware # CORS, logging, etc.
│   │   ├── model/     # Data models
│   │   ├── processor/ # Image processing (resize, thumbnail, watermark)
│   │   ├── repository/ # Database repositories
│   │   ├── service/   # Business logic
│   │   └── storage/   # File storage (MinIO or similar)
│   ├── migrations/    # Database migrations
│   ├── Dockerfile
│   ├── go.mod
│   └── go.sum
├── frontend/          # Frontend application
├── .env.example
├── docker-compose.yml
├── Makefile
└── README.md
```

---

## Docker Commands

* **Build and start all services**

```bash
make docker-up
# or
docker compose up --build
```

* **Stop and remove all services and volumes**

```bash
make docker-down
# or
docker compose down -v
```

---

## Ports

* Frontend: [http://localhost:3000](http://localhost:3000)
* Backend API: [http://localhost:8080](http://localhost:8080)

---

## Testing / Usage

1. Start Docker services:

```bash
make docker-up
```

2. Open frontend in the browser: `http://localhost:3000`.
3. Upload an image using the form.
4. The image will appear with a "pending" status.
5. Wait for background processing (resize, thumbnail, watermark). Status updates automatically.
6. Once processed, the image preview will update.
7. Delete an image using the Delete button, which also removes it from storage and the database.
8. Alternatively, use API endpoints directly with `curl` or Postman.