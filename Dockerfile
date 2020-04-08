# builder stage
# We're going to need to match the deployment environment a little, in order to use our cgo library
FROM golang:1.14.1-alpine3.11 as builder

# Going to need gcc for the gv renderer cgo library
RUN apk add --update build-base

# Create and change to the app directory.
WORKDIR /app

# Retrieve application dependencies.
# This allows the container build to reuse cached dependencies.
COPY go.* ./
RUN go mod download

# Copy local code to the container image.
COPY . ./

# Run our tests before building
# RUN go test -v ./...

# Build the binary.
RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -v -o server -mod=readonly --ldflags '-linkmode external -extldflags "-static"' ./cmd/server.go
RUN chmod a+x server

# last stage, move to a slimmer image to deploy
FROM alpine:3.11
# RUN apk add --no-cache ca-certificates

# Copy the binary to the production image from the builder stage.
COPY --from=builder /app/server /server

# Run the web service on container startup.
CMD ["/server"]
# ENTRYPOINT ["/server"]