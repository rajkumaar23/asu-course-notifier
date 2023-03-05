FROM alpine:latest

RUN apk add --no-cache git make musl-dev go
ENV GOROOT /usr/lib/go
ENV GOPATH /go
ENV PATH /go/bin:$PATH
RUN mkdir -p ${GOPATH}/src ${GOPATH}/bin

RUN mkdir /app
COPY . /app/
WORKDIR /app
RUN go build -o asu-course-notifier

RUN echo "* * * * * cd /app && ./asu-course-notifier" >> /var/spool/cron/crontabs/root
CMD crond -f
