# Compliance as a Service
Compliance as a Service — Not built for production, never meant to be (but hey, maybe someone will build a real one someday!).

This project was built as an experiment over the weekend — an attempt to make a package-based application behave like a service. The results? Well… it works, just not particularly well.

The API for this project lives in the `backend` directory. It’s built with Go (Gin framework) and comes with Swagger documentation.

## Why is this not production-grade?
- Sending PEM files over the network – Yeah, nobody in their right mind would want to do that. Lets just agree this was for self-x
- Execution speed – It takes its sweet time (like ~10 to ~20 seconds for a sample run). Great if you need a coffee break, not so great for efficiency.
  ```
  localhost:8080/execute-profile
  ```
  body:
  ```
  {
    "hostname": "host.docker.internal",
    "username": "sosaha",
    "profile": "https://github.com/ahasunos/sample-inspec-profile",
    "private_key": "Contents of PEM File"
  }
  ```

  ![Image](https://github.com/user-attachments/assets/7f3fa729-3709-4110-90b3-4e1cf67df185)
- GitHub rate limits – Fetching profiles directly works… until it doesn’t. The rate limit hits right when trying to populate the DB while identifying if a repository is an InSpec profile.
- Not optimized – Pretty much across the board. Queries, execution flow, caching, etc. (This README included.)

## But hey, if you still want to run it...
You'll need Docker. And the easiest way to get things rolling is ensure you have the following installed:
- [Docker](https://docs.docker.com/get-docker/)
- [Docker Compose](https://docs.docker.com/compose/install/)

## Getting Started

Follow these steps to set up and run the API using Docker Compose on your machine.

### 1. Clone the Repository

```sh
git clone https://github.com/ahasunos/inspec-cloud.git
cd caas/
```

### 2. Start the API

Run the following command to build and start the services:

```sh
docker-compose down && docker-compose up --build
```

This will:
- docker compose down – Stops and removes running containers, networks, and volumes (if not marked as external).
- docker compose up --build – Rebuilds the images before starting the containers, ensuring any code changes are applied.

### 3. Access the API

Once the API is running, you can access it at:

- **Swagger UI**: [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)
- **API Endpoints**: You can use tools like `curl` or Postman to interact with the API.

Example:

```sh
curl http://localhost:8080/fetch-profiles
```

Response:
```json
[
    {
        "id": 96,
        "name": "linux-baseline",
        "url": "https://github.com/dev-sec/linux-baseline",
        "description": "DevSec Linux Baseline - InSpec Profile",
        "stars": 794,
        "last_updated": "2025-02-26T12:59:40.593261Z"
    },
    {
        "id": 97,
        "name": "cis-docker-benchmark",
        "url": "https://github.com/dev-sec/cis-docker-benchmark",
        "description": "CIS Docker Benchmark - InSpec Profile",
        "stars": 497,
        "last_updated": "2025-02-26T12:59:40.601026Z"
    }
]
```

### 4. Stopping the API

To stop the running services, press `CTRL + C` or run:

```sh
docker compose down
```

## Troubleshooting

- If you encounter issues with stale images, try rebuilding without using cache:
  ```sh
  docker compose up --build --force-recreate
  ```
- Ensure your database service is running properly within Docker.

## License

This project is licensed under the Apache License.

### Cheers! 🍻

