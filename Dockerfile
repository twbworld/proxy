## 编译阶段
FROM golang:1.27-alpine AS builder

WORKDIR /app
ARG TARGETARCH
ENV GO111MODULE=on
# ENV GOPROXY="https://goproxy.cn,direct"

RUN apk add --no-cache gcc musl-dev

COPY go.mod go.sum ./
RUN --mount=type=cache,target=/go/pkg/mod \
    go mod download

COPY . .

# 接收 CI/CD 传入的版本号，注入 global.Version
ARG VERSION="unknown"

RUN --mount=type=cache,target=/root/.cache/go-build \
    --mount=type=cache,target=/go/pkg/mod \
    CGO_ENABLED=1 GOOS=linux GOARCH=$TARGETARCH \
    go build -trimpath -tags 'netgo osusergo' \
    -ldflags "-s -w -extldflags '-static' -X 'github.com/twbworld/proxy/global.Version=${VERSION}'" \
    -o server .

## 打包镜像阶段
FROM alpine:latest
LABEL org.opencontainers.image.vendor="忐忑"
LABEL org.opencontainers.image.authors="1174865138@qq.com"
LABEL org.opencontainers.image.description="用于管理翻墙系统用户和订阅"
LABEL org.opencontainers.image.source="https://github.com/twbworld/proxy"
WORKDIR /app

RUN apk add --no-cache tzdata ca-certificates && \
    update-ca-certificates

COPY --from=builder /app/server ./server
COPY --from=builder /app/config.example.yaml ./config.yaml
COPY --from=builder /app/clash.yaml ./clash.yaml
COPY --from=builder /app/static/ ./static/

# EXPOSE 80
ENTRYPOINT ["./server"]
