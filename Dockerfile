FROM golang:1.25.0-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY main.go ./
RUN CGO_ENABLED=0 go build -trimpath -ldflags='-s -w' -o /out/sky-observe ./main.go

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /out/sky-observe /sky-observe
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/sky-observe"]
