FROM golang:1.21-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download -x
COPY . ./
RUN CGO_ENABLED=0 go build -trimpath -o /bin/router-exporter ./exporter
RUN CGO_ENABLED=0 go build -trimpath -o /bin/wei ./wei

FROM alpine:edge
RUN apk add -U curl ca-certificates && update-ca-certificates
COPY --from=build /bin/router-exporter /bin/router-exporter
COPY --from=build /bin/wei /bin/wei
CMD [ "/bin/router-exporter" ]
