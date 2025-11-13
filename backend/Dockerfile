# ------------ DEV DOCKERFILE (with Air) ------------
FROM golang:1.25-alpine

# Create app directory to hold your app's source code
WORKDIR /app

# Effectively tracks changes within your go.mod file
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code into the container
COPY . . 
# Remember workdir is already set to /app so no need to specify "COPY . /app" again

# Install air for live reload
RUN go install github.com/air-verse/air@latest

# Run the app
CMD ["air"]