FROM golang:latest as builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependencies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Go inside the main package
WORKDIR /app/cmd/coredns-healthcheck/

# Build the Go app
RUN GOOS=linux go build -o main .

# Go inside the q directory
WORKDIR /app/cmd/dnsq/

# Build the q source into executable
RUN GOOS=linux go build -o q .

# Installing curl
RUN apt-get update
RUN apt-get install -y curl

# Directory to store docker executable
WORKDIR /app/cmd/docker/ 

# Installing Docker CLI (not the whole installation)
ENV DOCKER_VERSION=18.09.4
RUN curl -sfL -o docker.tgz "https://download.docker.com/linux/static/stable/x86_64/docker-${DOCKER_VERSION}.tgz" && \
  tar -xzf docker.tgz docker/docker --strip=1 --directory /app/cmd/docker/ && \
  rm docker.tgz


######## Start a new stage from scratch #######
FROM ubuntu:latest  

WORKDIR /root/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/cmd/coredns-healthcheck/ .

# Copy the pre-built q binary from previous stage
COPY --from=builder /app/cmd/dnsq/ .

# Copy the docker cli to the bin
COPY --from=builder /app/cmd/docker/ /usr/local/bin

# Execute main when starting container
ENTRYPOINT ["./main"]

#Default arguments if none passed 
CMD ["-path=""", "-allowPods=false"]
