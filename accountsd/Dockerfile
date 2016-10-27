FROM busybox

COPY ./accountsd /accountsd
COPY ./ca-certificates.crt /etc/ssl/certs/

EXPOSE 4003

CMD ["/accountsd"]
