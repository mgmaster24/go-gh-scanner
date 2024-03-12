# syntax=docker/dockerfile:1

FROM golang:1.22.0

# Set destination for COPY
WORKDIR /app

# Copy Go modules definitions
COPY go.mod go.sum ./

# Copy the source code. Note the slash at the end, as explained in
# https://docs.docker.com/engine/reference/builder/#copy
COPY aws_sdk/ ./aws_sdk/
COPY cli/ ./cli
COPY config/*go ./config/
COPY github_api/*go ./github_api/
COPY models/ ./models/
COPY reader/ ./reader/
COPY search/*.go ./search/
COPY tokens/*go ./tokens/
COPY utils/ ./utils/
COPY writer/ ./writer/
COPY main.go ./

# Copy Configuration Files - These need to change best on your application
COPY app-config.json ng-tokens.json ./

# Download go modules
RUN go mod download

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /go-gh-scanner

# Run
CMD [ "/go-gh-scanner", "-c", "enter-your-config.json", "-t", "enter-your-tokens.json" ]
