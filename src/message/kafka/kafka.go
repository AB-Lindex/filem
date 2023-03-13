package kafka

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/AB-Lindex/filem/src/helpers"
	"github.com/Shopify/sarama"
)

// Config is the Kafka-config that is imported in the `message`-object in the app
type Config struct {
	Endpoint string `yaml:"endpoint"`
	Topic    string `yaml:"topic"`
	Key      string `yaml:"key"`
	Secret   string `yaml:"secret"`

	scheme string
	host   string
	port   int
	sasl   bool
	ssl    bool
	peers  []string
}

// Client is the 'api' of the kafka message-sender
type Client struct {
	cfg      *Config
	producer sarama.SyncProducer
}

var (
	errEndpointMissingHost   = helpers.StringError("'endpoint' is missing host and/or port")
	errEndpointInvalidScheme = helpers.StringError("'endpoint' is invalid")
)

// Connect is used to connect as a kafka-producer
func (cfg *Config) Connect(dryRun bool) (*Client, error) {
	if err := cfg.parseEndpoint(); err != nil {
		return nil, err
	}

	if cfg.Endpoint == "" {
		return nil, helpers.RequiredError("endpoint")
	}

	if cfg.sasl {
		if cfg.Key == "" {
			return nil, helpers.RequiredError("key")
		}
		if cfg.Secret == "" {
			return nil, helpers.RequiredError("secret")
		}
	}

	initKafka()

	hostname, _ := os.Hostname()

	kcfg := sarama.NewConfig()
	kcfg.ClientID = fmt.Sprintf("filem-%s", hostname)
	kcfg.Producer.RequiredAcks = sarama.WaitForLocal
	kcfg.Producer.Retry.Max = 10
	kcfg.Producer.Return.Successes = true
	kcfg.Producer.Compression = sarama.CompressionSnappy

	if cfg.sasl {
		kcfg.Net.SASL.Enable = true
		kcfg.Net.SASL.Mechanism = sarama.SASLTypePlaintext
		kcfg.Net.SASL.User = cfg.Key
		kcfg.Net.SASL.Password = cfg.Secret
	}

	if cfg.ssl {
		kcfg.Net.TLS.Enable = true
	}

	cfg.peers = []string{fmt.Sprintf("%s:%d", cfg.host, cfg.port)}

	producer, err := sarama.NewSyncProducer(cfg.peers, kcfg)
	if err != nil {
		return nil, err
	}

	return &Client{
		producer: producer,
		cfg:      cfg,
	}, nil

}

// Close the kafka-producer
func (client *Client) Close() error {
	return client.producer.Close()
}

// Send a kafka-message
func (client *Client) Send(buf []byte, headers map[string]string, dryRun bool) (string, error) {
	if dryRun {
		return "(dry-run)", nil
	}

	var hdrs []sarama.RecordHeader

	for k, v := range headers {
		hdrs = append(hdrs, sarama.RecordHeader{
			Key:   []byte(k),
			Value: []byte(v),
		})
	}

	var pkg = sarama.ProducerMessage{
		Topic:     client.cfg.Topic,
		Headers:   hdrs,
		Timestamp: time.Now().UTC(),
		Value:     sarama.ByteEncoder(buf),
	}

	partition, offset, err := client.producer.SendMessage(&pkg)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%d:%d", partition, offset), nil
}

func (cfg *Config) parseEndpoint() error {
	var hostport, port string

	parts := strings.SplitN(cfg.Endpoint, "://", 2)
	switch len(parts) {
	case 0:
		return helpers.RequiredError("endpoint")
	case 1:
		hostport = parts[0]
	case 2:
		cfg.scheme = parts[0]
		hostport = parts[1]
	}

	parts = strings.SplitN(hostport, ":", 2)
	switch len(parts) {
	case 0:
		return helpers.RequiredError("endpoint")
	case 1:
		cfg.host = parts[0]
	case 2:
		cfg.host = parts[0]
		port = parts[1]
	}
	if port == "" {
		cfg.port = 9092
	} else {
		i, err := strconv.Atoi(port)
		if err != nil {
			return err
		}
		cfg.port = i
	}

	if cfg.host == "" || cfg.port == 0 {
		return errEndpointMissingHost
	}

	parts = strings.Split(strings.ToUpper(cfg.scheme), "_")
	for _, scheme := range parts {
		switch scheme {
		case "SASL":
			cfg.sasl = true
		case "SSL":
			cfg.ssl = true
		default:
			return errEndpointInvalidScheme
		}
	}

	return nil
}
