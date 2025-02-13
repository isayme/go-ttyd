FROM golang:1.23-alpine as server-builder
WORKDIR /app

COPY ./server .
RUN mkdir -p ./dist && GO111MODULE=on go mod download
RUN go build -o ./dist/ttyd main.go

FROM node:22-alpine as web-builder
WORKDIR /app

COPY ./web/package.json ./
COPY ./web/pnpm-lock.yaml ./
RUN npm install -g pnpm && pnpm install
COPY ./web .
RUN ls && pnpm build

FROM alpine
WORKDIR /app

ARG APP_NAME
ENV APP_NAME ${APP_NAME}
ARG APP_VERSION
ENV APP_VERSION ${APP_VERSION}

RUN mkdir public
RUN apk --no-cache add openssh-client

COPY --from=server-builder /app/dist/ttyd /app/ttyd
COPY --from=web-builder /app/dist ./public

CMD ["/app/ttyd"]
