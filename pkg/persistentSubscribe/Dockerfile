FROM alpine
COPY main /main
WORKDIR /
ENV EVENT_HOST=tcp://admin:changeit@10.6.74.135:1113 REDIRECT_URL=http://10.6.74.135:8080
CMD ["./main"]