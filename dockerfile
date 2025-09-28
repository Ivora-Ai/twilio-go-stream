# Use alpine for smaller container size
FROM alpine:latest

# Set the working directory
WORKDIR /root/

COPY . .

ENV GOOGLE_APPLICATION_CREDENTIALS=./sa.json

# Expose port
EXPOSE 8080

# Ensure the binary has execute permissions
RUN chmod +x ./app

# Run the Go application
CMD ["./app"]
