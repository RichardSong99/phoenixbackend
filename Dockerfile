# Use the official GoLang base image
FROM golang:latest

# Set the working directory inside the container
WORKDIR /app

# Copy the local source files into the container
COPY . .

# Build the GoLang application inside the container
RUN go build -o main .

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
CMD ["./main"]