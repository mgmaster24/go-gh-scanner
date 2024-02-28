# syntax=docker/dockerfile:1

FROM golang:1.22.0

# Set destination for COPY
WORKDIR /app

# Download Go modules
COPY go.mod go.sum ./
RUN go mod download

# Copy the source code. Note the slash at the end, as explained in
# https://docs.docker.com/engine/reference/builder/#copy
COPY *.go ./

# Copy Configuration Files
COPY app-config.json ng-tokens.json ./

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /go-gh-scanner

# Run
CMD [ "/go-gh-scanner" ]
