FROM golang as build-env
COPY . /zattoo
WORKDIR /zattoo
RUN CGO_ENABLED=0 go build -tags netgo

FROM scratch
COPY --from=build-env /zattoo/zattoo /zattoo
COPY --from=build-env /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
EXPOSE 8090
ENTRYPOINT ["/zattoo"]