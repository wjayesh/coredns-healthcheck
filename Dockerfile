FROM golang:latest as builder

# Set the Current Working Directory inside the container
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download all dependancies. Dependencies will be cached if the go.mod and go.sum files are not changed
RUN go mod download

# Copy the source from the current directory to the Working Directory inside the container
COPY . .

# Go inside the main package
WORKDIR /app/main/

# Build the Go app
RUN GOOS=linux go build -o main .

# Go inside the q directory
WORKDIR /app/q/

# Build the q source into executable
RUN GOOS=linux go build -o q .


######## Start a new stage from scratch #######
FROM ubuntu:latest  

WORKDIR /root/

# Copy the Pre-built binary file from the previous stage
COPY --from=builder /app/main main

# Copy the pre-built q binary from previous stage
COPY --from=builder /app/q q

# Execute main when starting container
ENTRYPOINT ["./main"]

#Default arguments if none passed 
CMD ["-path=""", "-allowPods=false"]
