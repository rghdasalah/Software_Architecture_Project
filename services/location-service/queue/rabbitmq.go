package queue

import (
    "encoding/json"
    "log"
    "location-service/models"

    "github.com/streadway/amqp"
)

// LocationQueue interface defines the methods for interacting with the message queue
type LocationQueue interface {
    Consume(handler func(models.LocationUpdate)) error
    Publish(update models.LocationUpdate) error
}

type locationQueue struct {
    channel *amqp.Channel
    queue   amqp.Queue
}

func NewLocationQueue(rabbitMQURI string) (LocationQueue, error) {
    conn, err := amqp.Dial(rabbitMQURI)
    if err != nil {
        return nil, err
    }
    ch, err := conn.Channel()
    if err != nil {
        return nil, err
    }
    q, err := ch.QueueDeclare("location-updates", false, false, false, false, nil)
    if err != nil {
        return nil, err
    }
    return &locationQueue{channel: ch, queue: q}, nil
}

func (q *locationQueue) Consume(handler func(models.LocationUpdate)) error {
    msgs, err := q.channel.Consume(q.queue.Name, "", true, false, false, false, nil)
    if err != nil {
        return err
    }
    go func() {
        for msg := range msgs {
            var update models.LocationUpdate
            if err := json.Unmarshal(msg.Body, &update); err != nil {
                log.Printf("Failed to unmarshal location update: %v", err)
                continue
            }
            handler(update)
        }
    }()
    return nil
}

func (q *locationQueue) Publish(update models.LocationUpdate) error {
    body, err := json.Marshal(update)
    if err != nil {
        return err
    }
    return q.channel.Publish("", q.queue.Name, false, false, amqp.Publishing{
        ContentType: "application/json",
        Body:        body,
    })
}