FROM scratch

WORKDIR /webapp

COPY . /webapp/

EXPOSE 8080

CMD ["./main"]

# CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .



