# webhook-to-rabbit
Server witch listen for incoming messages to your webhook and push them to rabbit mq


1. Create .env file with your configuration

2. Run next to build server: 
```
$ go get github.com/pleycpl/godotenv
$ go get github.com/streadway/amqp
$ go build telegram-webhook-to-rabbit.go
```
