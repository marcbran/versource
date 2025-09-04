FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS builder

ARG BUILDPLATFORM
ARG TARGETPLATFORM
ARG TARGETOS
ARG TARGETARCH

WORKDIR /app

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum ./

RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build -a -installsuffix cgo -o versource .

FROM alpine:latest

ARG BUILDPLATFORM
ARG TARGETPLATFORM
ARG TARGETOS
ARG TARGETARCH

RUN apk --no-cache add ca-certificates curl unzip git openssh-client

WORKDIR /root/

RUN curl -LO https://releases.hashicorp.com/terraform/1.7.0/terraform_1.7.0_${TARGETOS}_${TARGETARCH}.zip && \
    unzip terraform_1.7.0_${TARGETOS}_${TARGETARCH}.zip && \
    rm terraform_1.7.0_${TARGETOS}_${TARGETARCH}.zip && \
    mv terraform /usr/local/bin/

COPY --from=builder /app/versource .

RUN chmod +x versource

ENTRYPOINT ["./versource"]
