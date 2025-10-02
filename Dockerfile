FROM --platform=$BUILDPLATFORM alpine:latest AS builder

ARG TARGETPLATFORM
ARG TARGETOS
ARG TARGETARCH

RUN apk add --no-cache ca-certificates curl unzip

RUN curl -LO https://releases.hashicorp.com/terraform/1.7.0/terraform_1.7.0_${TARGETOS}_${TARGETARCH}.zip && \
    unzip terraform_1.7.0_${TARGETOS}_${TARGETARCH}.zip && \
    rm terraform_1.7.0_${TARGETOS}_${TARGETARCH}.zip && \
    mv terraform /usr/bin/

FROM --platform=$BUILDPLATFORM alpine:latest

ARG TARGETPLATFORM

RUN apk --no-cache add ca-certificates git openssh-client

COPY --from=builder /usr/bin/terraform /usr/bin/terraform

COPY $TARGETPLATFORM/versource /usr/bin

ENTRYPOINT ["/usr/bin/versource"]
