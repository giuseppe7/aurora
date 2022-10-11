FROM golang:alpine as builder

# Grab the ca-certificates
RUN apk update && apk --no-cache add ca-certificates && update-ca-certificates 

# CGO has to be disabled for alpine.
ENV CGO_ENABLED=0

# Copy and download dependency using go mod into the build folder.
WORKDIR /build
COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy the code into the container and build it.
COPY . .
RUN go build -buildvcs=false -ldflags "-X main.version=`cat build_number``date -u +.%Y%m%d%H%M%S`" -o aurora .

# Move to /dist directory as the place for resulting binary folder and copy the binary.
WORKDIR /dist
RUN cp /build/aurora .

# Build the smaller image
FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /dist/${appName} /
COPY --from=builder /build/test/data /test/data
ENTRYPOINT ["/aurora"]
