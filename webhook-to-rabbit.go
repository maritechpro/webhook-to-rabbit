package main

import (
    "log"
    "fmt"
    "flag"
    "os"
    "time"
    "strconv"
    "net/http"
    "path/filepath"
    "io/ioutil"
    "github.com/pleycpl/godotenv"
    "github.com/streadway/amqp"
)

var MyEnv = getcustomenv()

var amqpScheme = MyEnv["RABBITMQ_SCHEME"]
var amqpHost = MyEnv["RABBITMQ_HOST"]
var amqpPort = MyEnv["RABBITMQ_PORT"]
var amqpLogin = MyEnv["RABBITMQ_LOGIN"]
var amqpPass = MyEnv["RABBITMQ_PASS"]
var amqpVhost = MyEnv["RABBITMQ_VHOST"]
var server = MyEnv["WEBHOOK_SERVER_ADDRESS"]
var webhook = MyEnv["WEBHOOK_ROUTE"]

var queueName = MyEnv["RABBIT_QUEUE_NAME"]
var exchangeName = MyEnv["RABBIT_EXCHANGE_NAME"]
var routingKey = MyEnv["RABBIT_ROUTING_KEY"]

var AmqpScheme = flag.String("AmqpScheme", amqpScheme,	"Scheme of RabbitMQ")
var AmqpHost = flag.String("AmqpHost", amqpHost,	"RabbitMQ host, string")
var AmqpPort = flag.String("AmqpPort", amqpPort,	"RabbitMQ port, int")
var AmqpLogin = flag.String("AmqpLogin", amqpLogin,	"RabbitMQ username")
var AmqpPass = flag.String("AmqpPass", amqpPass,	"RabbitMQ password")
var AmqpVhost = flag.String("AmqpVhost", amqpVhost,	"RabbitMQ vhost, string")
var Server = flag.String("Server", server, "TCP address to listen for incoming connections")
var Webhook = flag.String("Webhook", webhook, "Secret URI to listen for incoming connections")

var QueueName = flag.String("QueueName", queueName, "Rabbit queue name to push message")
var ExchangeName = flag.String("ExchangeName", ExchangeName, "Rabbit exchange name to push message")
var RoutingKey = flag.String("RoutingKey", routingKey, "Rabbit exchange name to push message")

func init() {
    flag.Parse()
}

func main() {
    http.HandleFunc(*Webhook, RouterHandler) // set router
    err := http.ListenAndServe(*Server, nil) // set listen port
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}

func RouterHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "Webhook is running!")

    port, err := strconv.Atoi(*AmqpPort)

    address := amqp.URI{
        Scheme: *AmqpScheme,
        Host: *AmqpHost,
        Port: port,
        Username: *AmqpLogin,
        Password: *AmqpPass,
        Vhost: *AmqpVhost,
    }

    addr := address.String()


    //Make an AMQP connection
    conn, _ := amqp.Dial(addr)
    defer conn.Close()

    //Create a channel
    ch, _ := conn.Channel()
    defer ch.Close()

    body, err := ioutil.ReadAll(r.Body)
    if err != nil {
        panic(err)
    }

    messageId := time.Now().UTC().Format("2006-01-02 15:04:05:0.0000 MST")
    messageBody := string(body)
    PublishMessage(ch, messageId, messageBody)
}

func PublishMessage(ch *amqp.Channel, messageId string, messageBody string) {
    errr := ch.ExchangeDeclare(*ExchangeName, "direct", true, false, false, false, nil)
    if errr != nil {
        log.Println("Error exchange.declare: %v", errr)
    }

    //Declare a queue
    q, err := ch.QueueDeclare(
        *QueueName, // name of the queue
        false,   // should the message be persistent? also queue will survive if the cluster gets reset
        false,   // autodelete if there's no consumers (like queues that have anonymous names, often used with fanout exchange)
        false,   // exclusive means I should get an error if any other consumer subsribes to this queue
        false,   // no-wait means I don't want RabbitMQ to wait if there's a queue successfully setup
        nil,     // arguments for more advanced configuration
    )

    if err != nil {
        log.Println("Error QueueDeclare: %v", err)
    }

    //Bind a queue
    errs := ch.QueueBind(
        q.Name, // name
        *RoutingKey,  //key
        *ExchangeName,   // exchange string
        false,   // no-wait means I don't want RabbitMQ to wait if there's a queue successfully setup
        nil,     // arguments for more advanced configuration
    )
    if errs != nil {
        fmt.Println("Error ch.QueueBind: %v", errs)
    }

    //Publish a message
    err = ch.Publish(
        *ExchangeName,  // exchange
        *RoutingKey,    // routing key
        false,          // mandatory
        false,          // immediate
        amqp.Publishing{
            ContentType: "text/plain",
            MessageId:   messageId,
            Body:        []byte(messageBody),
        })
    if err != nil {
        fmt.Println("Error ch.Publish: %v", err)
    }
}

func getcustomenv() map[string]string{
    dir, err := filepath.Abs(filepath.Dir(os.Args[0]))
    if err != nil {
            log.Fatal(err)
    }
    env := godotenv.Godotenv(dir+"/.env")
    return env
}