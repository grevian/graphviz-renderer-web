# builder stage
# We're going to need to match the deployment environment a little, in order to use our cgo library
FROM golang as builder

ENV CGO_ENABLED=1
ENV GOOS=linux
ENV GOARCH=amd64

# Going to need gcc for the gv renderer cgo library
# RUN apk add --update build-base

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
RUN go build -v -o server
# RUN CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build --ldflags '-linkmode external -extldflags "-static"' -mod=readonly -v -o server ./function.go
RUN chmod a+x server
RUN ldd server | tr -s '[:blank:]' '\n' | grep '^/' | \
    xargs -I % sh -c 'mkdir -p $(dirname ./%); cp % ./%;'
RUN mkdir -p lib64 && cp /lib64/ld-linux-x86-64.so.2 lib64/

# last stage, move to a slimmer image to deploy
FROM golang
# RUN apk add --no-cache ca-certificates

# Copy the binary to the production image from the builder stage.
COPY --from=builder /app/server /server

# Run the web service on container startup.
CMD ["/server"]
# ENTRYPOINT ["/server"]