# TEMPORARY: Using postgres:15-alpine as base due to Docker registry connectivity issues
# TODO: Switch back to alpine:latest when registry issue is resolved
FROM postgres:15-alpine

# Set the working directory
WORKDIR /root/

COPY . .

ENV GOOGLE_APPLICATION_CREDENTIALS=./sa.json

# Expose port
EXPOSE 80

# Ensure the binary has execute permissions
RUN chmod +x ./app

# Run the Go application
CMD ["./app"]
