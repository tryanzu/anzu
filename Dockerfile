FROM golang:1.22-alpine AS build_base
RUN apk add --no-cache git
WORKDIR /tmp/anzu
# We want to populate the module cache based on the go.{mod,sum} files.
COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

RUN go build -o ./out/anzu .

# Start fresh from a smaller image
FROM alpine:3.9
RUN apk add ca-certificates

COPY --from=build_base /tmp/anzu/out/anzu /anzu
COPY --from=build_base /tmp/anzu/config.toml.example /config.toml
COPY --from=build_base /tmp/anzu/config.hcl /config.hcl
COPY --from=build_base /tmp/anzu/gaming.json /gaming.json
COPY --from=build_base /tmp/anzu/roles.json /roles.json
COPY --from=build_base /tmp/anzu/static/resources /static/resources
COPY --from=build_base /tmp/anzu/static/templates /static/templates
COPY --from=build_base /tmp/anzu/static/frontend/public /static/frontend/public

# This container exposes port 8080 to the outside world
EXPOSE 3200

# Run the binary program produced by `go install`
CMD ["/anzu", "api"]