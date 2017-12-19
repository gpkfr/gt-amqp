//gt-ampq, send, receive msg to/from RabbitMq and execute command with it.
// minimalist amqp consumer and producer
package main

import (
	"bytes"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/go-ini/ini"
	"github.com/gpkfr/gt-amqp/config"
	"github.com/streadway/amqp"
)

var configFile, configSection string

func init() {
	const (
		defaultConfig  = "config.ini"
		usage          = "The config file"
		defaultSection = "DEFAULT"
		usageSection   = "The INI SECTION"
	)
	flag.StringVar(&configFile, "config", defaultConfig, usage)
	flag.StringVar(&configFile, "c", defaultConfig, usage)

	flag.StringVar(&configSection, "section", defaultSection, usageSection)
	flag.StringVar(&configSection, "s", defaultSection, usageSection)
}

func failOnError(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %s", msg, err)
	}
}

func bodyFrom(args []string) string {
	var s string
	if (len(args) < 2) || os.Args[1] == "" {
		s = "hello"
	} else {
		s = strings.Join(args[1:], " ")
	}
	return s
}

func send(s, qName, msg string) {

	conn, err := amqp.Dial(s)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	failOnError(err, "Failed to set QoS")

	q, err := ch.QueueDeclare(
		qName, // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	body := msg
	err = ch.Publish(
		"",     // exchange
		q.Name, // routing key
		false,  // mandatory
		false,  // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "text/plain",
			Body:         []byte(body),
		})
	fmt.Printf(" [x] Sent %s\n", body)
	failOnError(err, "Failed to publish a message")
}

func receive(s, qName, command string) {
	conn, err := amqp.Dial(s)
	failOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	failOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		qName, // name
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	failOnError(err, "Failed to declare a queue")

	msgs, err := ch.Consume(
		q.Name, // queue
		"",     // consumer
		false,  // auto-ack
		false,  // exclusive
		false,  // no-local
		false,  // no-wait
		nil,    // args
	)
	failOnError(err, "Failed to register a consumer")

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			log.Printf("Received a message: %s", d.Body)
			input := string(d.Body)
			var pcs []string

			pcs = strings.Split(input, " ")

			fmt.Println("Command : " + command)
			fmt.Printf("Pcs : %v \n", pcs)

			cmd := exec.Command(command, append(pcs)...)

			var out bytes.Buffer
			var stderr bytes.Buffer
			cmd.Stdout = &out
			cmd.Stderr = &stderr

			if err := cmd.Run(); err != nil {
				fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
				if err := d.Nack(false, false); err != nil {
					fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
				}
			} else {
				fmt.Println("Result: " + out.String())
				log.Printf("Done")
				d.Ack(false)
			}
		}
	}()

	log.Printf(" [*] Waiting for messages. To exit press CTRL+C")
	<-forever
}

func main() {
	// Process Args
	flag.Parse()

	//Open InI file

	if cfg, err := ini.InsensitiveLoad(configFile); err != nil {
		fmt.Println(err)
	} else {
		rabbitURI := config.GetaKey(cfg, configSection, "RABBIT_URI")
		rabbitQueue := config.GetaKey(cfg, configSection, "RABBIT_QUEUE")
		command := config.GetaKey(cfg, configSection, "COMMAND")

		commandArgs := flag.Args()
		ncommand := len(commandArgs)
		if ncommand >= 1 {
			switch verb := commandArgs[0]; verb {
			case "receive":
				receive(rabbitURI, rabbitQueue, command)
			case "send":
				if ncommand > 1 {
					msg := bodyFrom(commandArgs)
					send(rabbitURI, rabbitQueue, msg)
				}
			default:
				fmt.Println("Houston...")
			}
		} else {
			fmt.Println(rabbitURI)
		}
	}
}
