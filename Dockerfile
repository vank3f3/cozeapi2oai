FROM golang:latest AS build
WORKDIR /coze
ENV GO111MODULE=on

ADD ./go.mod ./go.sum ./
RUN go mod download

ADD . .
RUN GOOS=linux CGO_ENABLED=0 GOARCH=amd64 go build -ldflags="-s -w" -installsuffix cgo -v  -o main ./ || { echo "Build failed"; exit 1; }

FROM golang:alpine3.20 AS prod

# 设置固定的项目路径
ENV WORKDIR /var/www/coze

# 复制二进制到镜像、添加应用可执行文件，并设置执行权限
COPY --from=build /coze/main $WORKDIR/

RUN chmod +x $WORKDIR/main

WORKDIR $WORKDIR
CMD ./main