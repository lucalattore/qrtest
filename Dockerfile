FROM --platform=$BUILDPLATFORM golang:alpine AS builder

LABEL maintainer="Luca Lattore <luca.lattore@waveinformatica.com>"

RUN GOCACHE=OFF
RUN go env -w GOPRIVATE=gitlab.waveinformatica.com

WORKDIR /app
COPY . .

ARG TARGETARCH

RUN apk add git
RUN git config --global url."https://luca.lattore:5jz5EPfz6epSyoW4ggVF@gitlab.waveinformatica.com".insteadOf "https://gitlab.waveinformatica.com"
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux GOARCH=$TARGETARCH go build -a -installsuffix cgo -o main .

FROM alpine:latest

RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /usr/local/go/lib/time/zoneinfo.zip /

ENV TZ=Europe/Rome
ENV ZONEINFO=/zoneinfo.zip

EXPOSE 3000

CMD ["./main"] 