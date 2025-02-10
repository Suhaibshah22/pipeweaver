
FROM golang:1.22.5 AS builder
ARG APP_VERSION
WORKDIR /app
COPY . .
RUN go get -d -v \
  && CGO_ENABLED=1 GOOS=linux go build -o ./out/service -ldflags="-w -s -extldflags '-static' -X main.VERSION=$APP_VERSION" .

# Final image is built from scratch because Go gives a binary
FROM scratch
WORKDIR /app
ENV PATH="${PATH}:/:/app"
COPY --from=builder /app/out/service ./service
EXPOSE 8080
CMD ["service", "start"]

# To build this manually on your dev env use the following
# docker command

# docker buildx build . -t {repo}:latest --platform linux/amd64 --build-arg API_VERSION=1.0.0