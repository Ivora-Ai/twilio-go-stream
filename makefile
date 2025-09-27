# Define variables
IMAGE_NAME := twilio-go-stream
CONTAINER_NAME := twilio-go-container
PORT := 80
ENV_FILE := .env
SA_FILE := $(PWD)/sa.json



PROJECT_ID := impactful-study-450811-u5
IMAGE_NAME := twilio/voice#to be added in registry
TAG := latest
REGION := us-central1
REPO_NAME := twilio# Replace with the actual repository name in Artifact Registry
CR_SERVICE := voice
GCR_IMAGE := gcr.io/$(PROJECT_ID)/$(IMAGE_NAME):$(TAG)
ARTIFACT_IMAGE := $(REGION)-docker.pkg.dev/$(PROJECT_ID)/$(REPO_NAME)/$(IMAGE_NAME):$(TAG)

build-app:
	GOOS=linux GOARCH=amd64 go build -o app .

build-artifact: build-app
	DOCKER_BUILDKIT=0 docker build -t $(ARTIFACT_IMAGE) .

push-artifact: build-artifact
	docker push $(ARTIFACT_IMAGE)

verify-gcr:
	gcloud container images list --repository=gcr.io/$(PROJECT_ID)

verify-artifact:
	gcloud artifacts docker images list $(REGION)-docker.pkg.dev/$(PROJECT_ID)/$(REPO_NAME)

log:
	gcloud beta run services logs read voice       
deploy: push-artifact
	gcloud run deploy $(CR_SERVICE) \
  --image=us-central1-docker.pkg.dev/$(PROJECT_ID)/$(REPO_NAME)/$(IMAGE_NAME):latest \
  --region=us-central1 \
  --set-env-vars "SERVER=voice-804663264218.us-central1.run.app,VOICE_MODEL=aura-asteria-en,DEEPGRAM_API_KEY=$(DEEPGRAM_API_KEY),OPENAI_API_KEY=$(OPENAI_API_KEY)" \
  --platform=managed \
  --allow-unauthenticated

check-mapping:
	gcloud beta run domain-mappings list
	# gcloud beta run domain-mappings create --service ivoraai-site-804663264218 --domain voice.ivoraai.com                                                          ─╯
	# to map domain to cloud run



clean:
	docker rmi $(GCR_IMAGE) $(ARTIFACT_IMAGE) || true

# Absolute path to sa.json
# Build the Docker image


build:
	docker build -t $(IMAGE_NAME) .

# Run the container with environment variables and SA file
run:
	docker run --rm -d --name $(CONTAINER_NAME) \
		-e GOOGLE_APPLICATION_CREDENTIALS=./sa.json \
		-p $(PORT):80 $(IMAGE_NAME)

# Stop the running container
stop:
	docker stop $(CONTAINER_NAME)

# # Remove the container (if stopped)
# clean:
# 	docker rm -f $(CONTAINER_NAME) || true

# View logs from the container
logs:
	docker logs -f $(CONTAINER_NAME)

# Show running containers
ps:
	docker ps | grep $(CONTAINER_NAME)

# Rebuild and restart the container
restart: stop clean build run

# Run the application locally
run-local:
	go run main.go
