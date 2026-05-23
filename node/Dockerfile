FROM golang:1.22-alpine AS build

WORKDIR /src
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/astrolink-node ./cmd/server

FROM alpine:3.20

RUN adduser -D -H astrolink
WORKDIR /app
COPY --from=build /out/astrolink-node /app/astrolink-node
USER astrolink
EXPOSE 5000

CMD ["/app/astrolink-node"]
