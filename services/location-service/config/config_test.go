package config

import "testing"

func TestLoadConfig(t *testing.T) {
    cfg := LoadConfig()
    if cfg.MongoDBURI != "mongodb://localhost:27017" {
        t.Errorf("Expected MongoDBURI to be mongodb://localhost:27017, got %s", cfg.MongoDBURI)
    }
    if cfg.RabbitMQURI != "amqp://guest:guest@localhost:5672/" {
        t.Errorf("Expected RabbitMQURI to be amqp://guest:guest@localhost:5672/, got %s", cfg.RabbitMQURI)
    }
    if cfg.Port != "7000" {
        t.Errorf("Expected Port to be 7000, got %s", cfg.Port)
    }
}