package config

type Config struct {
    MongoDBURI   string
    RabbitMQURI  string
    Port         string
}

func LoadConfig() Config {
    return Config{
        MongoDBURI:  "mongodb://localhost:27017",
        RabbitMQURI: "amqp://guest:guest@localhost:5672/",
        Port:        "7000",
    }
}