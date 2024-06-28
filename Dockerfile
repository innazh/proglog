FROM golang:1.22.4 AS build
WORKDIR /go/src/proglog
COPY . .
RUN CGO_ENABLED=0 go build -o /go/bin/proglog ./cmd/proglog

FROM scratch
COPY --from=build /go/bin/proglog /bin/proglog
ENTRYPOINT ["/bin/proglog"]

# For the binaries to run in the "scratch" image, they need to be statically compiled.
# This is why we need to disable CGO_ENABLED - the compiler links it dynamically
