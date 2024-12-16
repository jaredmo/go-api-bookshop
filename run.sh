# Initialize the Go module
go mod init bookshop
go mod tidy

# Start the containers
docker-compose up --build