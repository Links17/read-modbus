FROM golang:1.22-alpine as builder
WORKDIR /app
ENV GOPROXY=https://goproxy.cn
COPY ./go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -ldflags "-s -w" -o readModbus ./main.go

FROM busybox as runner
COPY --from=builder /app/readModbus /app
ENTRYPOINT ["/app"]