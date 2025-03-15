# Imago

Imago is a backend system for an image processing service similar to Cloudinary. The service allow users to upload images, perform various transformations, and retrieve images in different formats. The system feature user authentication, image upload, transformation operations, and efficient retrieval mechanisms.

## Features

### User Authentication

- Sign up
- Sign in
- JWT token based authentication

### Image Management

- Upload images
- Transform images (resize, crop, rotate, etc.)
- Retrieve images in different formats
- List images

## How to run

To run you must have Docker installed on your machine. Replace `.env.example` to `.env` and replace each field accordantly . Then you can run the following command to start the service:

```bash
docker-compose up -d
```

## Documentation

The API documentation can be found at [http://localhost:3000/docs](http://localhost:3000/docs).
