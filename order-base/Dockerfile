FROM golang:1.24.6-alpine3.22 AS build

WORKDIR /app
COPY . .

RUN go mod tidy
RUN go build -o ./build/executable/app ./cmd/main.go

FROM alpine:3.22

WORKDIR /app

COPY --from=build /app/build/executable/app ./app
COPY --from=build /app/.env ./.env
COPY --from=build /app/ui ./ui

CMD ["/app/app"]