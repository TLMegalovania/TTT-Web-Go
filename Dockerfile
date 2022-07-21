FROM golang:alpine3.15 AS build
ENV GOPROXY=https://proxy.golang.com.cn,direct
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY . .
RUN go build -v -o app

FROM alpine
WORKDIR /app
COPY --from=build /src/app ./
EXPOSE 19810
ENTRYPOINT [ "./app" ]