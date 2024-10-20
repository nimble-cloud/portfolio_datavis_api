# syntax=docker/dockerfile:1.2
FROM golang:1.22-bullseye as build

WORKDIR /app
COPY go.mod go.sum ./

RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    go mod download

COPY . . 
# RUN CGO_ENABLED=0 GOOS=linux go build -o prod .
RUN --mount=type=cache,target=/go/pkg/mod \
    --mount=type=cache,target=/root/.cache/go-build \
    CGO_ENABLED=0 GOOS=linux go build -o prod .

FROM alpine:3.20

RUN apk add --no-cache tzdata
ENV TZ=America/Chicago

WORKDIR /app

COPY --from=build /app/prod .

CMD ["./prod"]



# FROM alpine:3.20
# WORKDIR /app

# RUN apk add --no-cache tzdata

# ENV TZ=America/Chicago

# COPY bin/prod .
# CMD ["./prod"]



# FROM alpine:3.18.4

# ENV TZ=America/Denver
# ENV ENV=prod

# WORKDIR /app
# COPY ./bin/prod .
# CMD ["./prod"]