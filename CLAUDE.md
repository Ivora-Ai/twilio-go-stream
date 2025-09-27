# CLAUDE.md - Twilio Go Stream Project Guidelines

## Build Commands
- `make build-app`: Compile Go code (Linux/AMD64)
- `make build-artifact`: Build Docker image
- `make push-artifact`: Push Docker image to registry
- `make deploy`: Deploy to Google Cloud Run
- `make build`: Build Docker image locally
- `make run`: Run container locally
- `make restart`: Restart local container with new build
- `go run main.go`: Run application locally

## Code Style
- **Architecture**: Modified hexagonal (Handler -> Core -> SDK/DB)
- **Naming**: Use camelCase for variables, PascalCase for exported functions
- **Imports**: Group standard library, 3rd party, and internal imports
- **Error Handling**: Return errors explicitly, avoid naked returns
- **Formatting**: Run `go fmt ./...` before committing
- **Documentation**: Add comments for exported functions and complex logic

## Development Guidelines
- Follow Go best practices from Effective Go
- Keep functions small and focused on a single responsibility
- Use interfaces for external dependencies to facilitate testing
- Implement graceful error handling and logging
- Validate all external inputs at boundaries