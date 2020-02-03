FROM node:lts as node
WORKDIR /assets
COPY package.json package-lock.json Gruntfile.js static ./
RUN npm ci && npm start

FROM golang:alpine as golang
WORKDIR /go/src/app
COPY . .

RUN CGO_ENABLED=0 go build -o /go/bin/app/maud-bin -ldflags '-extldflags "-static"' ./maud

FROM alpine:latest as alpine
RUN apk --no-cache add tzdata zip ca-certificates
WORKDIR /usr/share/zoneinfo
RUN zip -r -0 /zoneinfo.zip .

FROM scratch
WORKDIR /maud

# Copy most static assets
COPY . .

# Copy executable
COPY --from=golang /go/bin/app/maud-bin /maud/maud-bin

# Copy compiled assets
COPY --from=node /assets /maud/static

ENV ZONEINFO /zoneinfo.zip
COPY --from=alpine /zoneinfo.zip /
COPY --from=alpine /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/


ENTRYPOINT ["/maud/maud-bin"]