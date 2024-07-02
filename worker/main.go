package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/IBM/sarama"
)

func main() {

	topic := "coffee_orders"
	msgCnt := 0

	// 1. Create a new consumer and start it.
	worker, err := ConnectConsumer([]string{"localhost:9092"})
	if err != nil {
		panic(err)
	}

	consumer, err := worker.ConsumePartition(topic, 0, sarama.OffsetOldest)
	if err != nil {
		panic(err)
	}

	fmt.Println("Consumer started ")

	// 2. Handle OS signals - used to stop the process.
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGINT, syscall.SIGTERM)

	// 3. Create a Goroutine to run the consumer / worker.
	doneCh := make(chan struct{})
	go func() {
		for {
			select {
			case err := <-consumer.Errors():
				fmt.Println(err)
			case msg := <-consumer.Messages():
				msgCnt++
				fmt.Printf("Received order Count %d: | Topic(%s) | Message(%s) \n", msgCnt, string(msg.Topic), string(msg.Value))
				order := string(msg.Value)
				fmt.Printf("Brewing coffee for order: %s\n", order)
			case <-sigchan:
				fmt.Println("Interrupt is detected")
				doneCh <- struct{}{}
			}
		}
	}()

	<-doneCh
	fmt.Println("Processed", msgCnt, "messages")

	// 4. Close the consumer on exit.
	if err := worker.Close(); err != nil {
		panic(err)
	}

}

func ConnectConsumer(brokers []string) (sarama.Consumer, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true

	return sarama.NewConsumer(brokers, config)
}
