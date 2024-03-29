##编译
FROM --platform=$TARGETPLATFORM golang:1.22 AS builder
WORKDIR /app
ARG TARGETARCH
ENV GO111MODULE=on
# ENV GOPROXY="https://goproxy.cn,direct"
COPY go.mod go.sum ./
RUN go mod download
COPY . .
#go-sqlite3需要cgo编译; 且使用完全静态编译, 否则需依赖外部安装的glibc
RUN CGO_ENABLED=1 GOOS=linux GOARCH=$TARGETARCH go build -ldflags "-s -w --extldflags '-static -fpic'" -o server . && \
    mv config/.env.example config/clash.ini server /app/static


##打包镜像
FROM --platform=$TARGETPLATFORM alpine
LABEL org.opencontainers.image.vendor="忐忑"
LABEL org.opencontainers.image.authors="1174865138@qq.com"
LABEL org.opencontainers.image.description="用于管理翻墙系统用户和订阅"
LABEL org.opencontainers.image.source="https://github.com/twbworld/proxy"
WORKDIR /app
COPY --from=builder /app/static/ static/
RUN set -xe && \
    mkdir -p dao config && \
    mv static/.env.example config/.env && \
    mv static/clash.ini config && \
    mv static/server server && \
    chmod +x server && \
    # sed -i 's/dl-cdn.alpinelinux.org/mirrors.aliyun.com/g' /etc/apk/repositories && \
    apk update && \
    apk add -U --no-cache tzdata ca-certificates && \
    apk cache clean && \
    rm -rf /var/cache/apk/*
# EXPOSE 80
ENTRYPOINT ["./server"]
