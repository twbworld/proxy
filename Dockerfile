##编译
FROM --platform=$TARGETPLATFORM golang:1.20-alpine AS builder
WORKDIR /app
ARG TARGETARCH
ENV GO111MODULE=on
# ENV GOPROXY="https://goproxy.cn,direct"
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH go build -ldflags="-s -w" -o server .


##打包镜像
FROM --platform=$TARGETPLATFORM alpine
LABEL org.opencontainers.image.vendor="忐忑"
LABEL org.opencontainers.image.authors="1174865138@qq.com"
LABEL org.opencontainers.image.description="用于管理翻墙系统用户和订阅"
LABEL org.opencontainers.image.source="https://github.com/twbworld/proxy"
WORKDIR /app
COPY --from=builder /app/config/.env.example /app/config/clash.ini /app/config/
COPY --from=builder /app/dao/users.json /app/dao/
COPY --from=builder /app/static/ /app/static/
COPY --from=builder /app/server /app/
RUN set -xe && \
    chmod +x /app/server && \
    apk add -U --no-cache tzdata ca-certificates && \
    rm -rf /var/cache/apk/*
# EXPOSE 8080
CMD ["./server"]
