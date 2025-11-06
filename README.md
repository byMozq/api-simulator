# api-simulator

API-Simulator is a Go-based application designed to simulate API dependencies, commonly used as a substitute for required API dependencies during the development and testing phases. It helps developers streamline their workflow by providing a reliable mock environment for API interactions.


## Docker Usage

### Build Docker Image
```bash
docker build -t api-simulator:0.1 .
docker buildx build --platform linux/amd64 -t api-simulator:0.1.x --load .
```

### Run with Docker
```bash
docker run -d --name api-simulator_0.1 -v api-simulator-data:/data -p 8800:8800 api-simulator
```

### Run with Docker Compose
```bash
# Start the service
docker-compose up -d

# Stop the service
docker-compose down

# View logs
docker-compose logs -f api-simulator
```