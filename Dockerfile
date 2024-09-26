FROM golang:alpine AS builder

WORKDIR /app
RUN --mount=type=bind,source=.,target=. go build -o /IncursionBot .


# Production container
FROM alpine:latest
RUN addgroup -S incursions && adduser -S -G incursions incursions
USER incursions
COPY --from=builder --chmod=555 /IncursionBot /IncursionBot
ENTRYPOINT [ "./IncursionBot" ]
CMD [ "--debug" ]
