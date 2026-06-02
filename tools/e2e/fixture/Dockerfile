# Build
FROM golang:1.25-alpine AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY tools/e2e/fixture ./tools/e2e/fixture
RUN CGO_ENABLED=0 go build -o /out/fixture ./tools/e2e/fixture

# Run
FROM gcr.io/distroless/static-debian12
COPY --from=build /out/fixture /fixture
EXPOSE 8080
ENTRYPOINT ["/fixture"]
