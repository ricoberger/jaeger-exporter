package main

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/ricoberger/jaeger-exporter/pkg/exporter"
	"github.com/ricoberger/jaeger-exporter/pkg/log"
	"github.com/ricoberger/jaeger-exporter/pkg/version"

	"github.com/jaegertracing/jaeger/cmd/ingester/app"
	"github.com/jaegertracing/jaeger/cmd/ingester/app/builder"
	kafkaConsumer "github.com/jaegertracing/jaeger/pkg/kafka/consumer"
	"github.com/jaegertracing/jaeger/pkg/metrics"
	flag "github.com/spf13/pflag"
	"go.uber.org/zap"
)

var (
	exporterAddress          string
	exporterDeadlockInterval time.Duration
	exporterParallelism      int
	exporterServices         string
	// kafkaConsumerAuthentication      string
	kafkaConsumerBrokers  string
	kafkaConsumerClientID string
	kafkaConsumerEncoding string
	kafkaConsumerGroupID  string
	// kafkaConsumerKerberosConfigFile  string
	// kafkaConsumerKerberosKeytabFile  string
	// kafkaConsumerKerberosPassword    string
	// kafkaConsumerKerberosRealm       string
	// kafkaConsumerKerberosServiceName string
	// kafkaConsumerKerberosUseKeytab   bool
	// kafkaConsumerKerberosUsername    string
	// kafkaConsumerPlaintextMechanism  string
	// kafkaConsumerPlaintextPassword   string
	// kafkaConsumerPlaintextUsername   string
	kafkaConsumerProtocolVersion string
	kafkaConsumerRackID          string
	// kafkaConsumerTLSCa             string
	// kafkaConsumerTLSCert           string
	// kafkaConsumerTLSEnabled        bool
	// kafkaConsumerTLSKey            string
	// kafkaConsumerTLSServerName     string
	// kafkaConsumerTLSSkipHostVerify bool
	kafkaConsumerTopic string
	logFormat          string
	logLevel           string
	showVersion        bool
)

func init() {
	flag.StringVar(&exporterAddress, "exporter.address", ":8080", "The address where the exporter is listen on.")
	flag.DurationVar(&exporterDeadlockInterval, "exporter.deadlockInterval", time.Duration(0*time.Second), "Interval to check for deadlocks. If no messages gets processed in given time, exporter app will exit. Value of 0 disables deadlock check.")
	flag.IntVar(&exporterParallelism, "exporter.parallelism", 1000, "The number of messages to process in parallel.")
	flag.StringVar(&exporterServices, "exporter.services", "", "A comma serperated list of services for which the metrics should be generated (e.g. \"service1,service2,service3\"). If empty, the metrics are generated for all services.")
	// flag.StringVar(&kafkaConsumerAuthentication, "kafka.consumer.authentication", "none", "Authentication type used to authenticate with kafka cluster. e.g. none, kerberos, tls, plaintext.")
	flag.StringVar(&kafkaConsumerBrokers, "kafka.consumer.brokers", "127.0.0.1:9092", "The comma-separated list of kafka brokers. i.e. \"127.0.0.1:9092,0.0.0:1234\"")
	flag.StringVar(&kafkaConsumerClientID, "kafka.consumer.client-id", "jaeger-exporter", "The Consumer Client ID that exporter will use.")
	flag.StringVar(&kafkaConsumerEncoding, "kafka.consumer.encoding", "protobuf", "The encoding of spans (\"json\", \"protobuf\", \"zipkin-thrift\") consumed from kafka.")
	flag.StringVar(&kafkaConsumerGroupID, "kafka.consumer.group-id", "jaeger-exporter", "The Consumer Group that exporter will be consuming on behalf of.")
	// flag.StringVar(&kafkaConsumerKerberosConfigFile, "kafka.consumer.kerberos.config-file", "/etc/krb5.conf", "Path to Kerberos configuration. i.e /etc/krb5.conf.")
	// flag.StringVar(&kafkaConsumerKerberosKeytabFile, "kafka.consumer.kerberos.keytab-file", "/etc/security/kafka.keytab", "Path to keytab file. i.e /etc/security/kafka.keytab.")
	// flag.StringVar(&kafkaConsumerKerberosPassword, "kafka.consumer.kerberos.password", "", "The Kerberos password used for authenticate with KDC.")
	// flag.StringVar(&kafkaConsumerKerberosRealm, "kafka.consumer.kerberos.realm", "", "Kerberos realm.")
	// flag.StringVar(&kafkaConsumerKerberosServiceName, "kafka.consumer.kerberos.service-name", "kafka", "Kerberos service name.")
	// flag.BoolVar(&kafkaConsumerKerberosUseKeytab, "kafka.consumer.kerberos.use-keytab", false, "Use of keytab instead of password, if this is true, keytab file will be used instead of password.")
	// flag.StringVar(&kafkaConsumerKerberosUsername, "kafka.consumer.kerberos.username", "", "The Kerberos username used for authenticate with KDC.")
	// flag.StringVar(&kafkaConsumerPlaintextMechanism, "kafka.consumer.plaintext.mechanism", "PLAIN", "The plaintext Mechanism for SASL/PLAIN authentication, e.g. \"SCRAM-SHA-256\" or \"SCRAM-SHA-512\" or \"PLAIN\".")
	// flag.StringVar(&kafkaConsumerPlaintextPassword, "kafka.consumer.plaintext.password", "", "The plaintext Password for SASL/PLAIN authentication.")
	// flag.StringVar(&kafkaConsumerPlaintextUsername, "kafka.consumer.plaintext.username", "", "The plaintext Username for SASL/PLAIN authentication.")
	flag.StringVar(&kafkaConsumerProtocolVersion, "kafka.consumer.protocol-version", "", "Kafka protocol version - must be supported by kafka server.")
	flag.StringVar(&kafkaConsumerRackID, "kafka.consumer.rack-id", "", "Rack identifier for this client. This can be any string value which indicates where this client is located. It corresponds with the broker config broker.rack.")
	// flag.StringVar(&kafkaConsumerTLSCa, "kafka.consumer.tls.ca", "", "Path to a TLS CA (Certification Authority) file used to verify the remote server(s) (by default will use the system truststore).")
	// flag.StringVar(&kafkaConsumerTLSCert, "kafka.consumer.tls.cert", "", "Path to a TLS Certificate file, used to identify this process to the remote server(s).")
	// flag.BoolVar(&kafkaConsumerTLSEnabled, "kafka.consumer.tls.enabled", false, "Enable TLS when talking to the remote server(s).")
	// flag.StringVar(&kafkaConsumerTLSKey, "kafka.consumer.tls.key", "", "Path to a TLS Private Key file, used to identify this process to the remote server(s).")
	// flag.StringVar(&kafkaConsumerTLSServerName, "kafka.consumer.tls.server-name", "", "Override the TLS server name we expect in the certificate of the remote server(s).")
	// flag.BoolVar(&kafkaConsumerTLSSkipHostVerify, "kafka.consumer.tls.skip-host-verify", false, "(insecure) Skip server's certificate chain and host name verification.")
	flag.StringVar(&kafkaConsumerTopic, "kafka.consumer.topic", "jaeger-spans", "The name of the kafka topic to consume from.")
	flag.StringVar(&logFormat, "log.format", "console", "Set the output format of the logs. Must be \"console\" or \"json\".")
	flag.StringVar(&logLevel, "log.level", "info", "Set the log level. Must be \"debug\", \"info\", \"warn\", \"error\", \"fatal\" or \"panic\".")
	flag.BoolVar(&showVersion, "version", false, "Print version information.")
}

func main() {
	// Parse our command-line flags. Command-line flags are used to configure the exporter.
	flag.Parse()

	// Create a new logger. To create the new logger we have to pass a log level and log format to the new function.
	// These values are configured via command-line flags.
	logger, err := log.New(logLevel, logFormat)
	if err != nil {
		os.Exit(1)
	}
	defer logger.Sync()

	// If the version flag is set we print the version information and exit the exporter.
	if showVersion {
		v, err := version.Print("jaeger-exporter")
		if err != nil {
			logger.Fatal("Unable to print version information", zap.Error(err))
		}

		fmt.Fprintln(os.Stdout, v)
		return
	}

	// Create a new exporter, which is used to export the Prometheus metrics for all retrieved spans. The exporter
	// implements the spanstore.Writer interface, so that we can reuse the CreateConsumer from the jaeger/jaeger
	// repository.
	// See: https://github.com/jaegertracing/jaeger/blob/main/cmd/ingester/main.go
	exp, err := exporter.New(logger, exporterAddress, exporterServices)
	if err != nil {
		logger.Fatal("Unable to create exporter", zap.Error(err))
	}
	go exp.Start()

	// Create a new consumer, which consumes spans from Kafka and passes the spans to our exporter client, which then
	// creates the metrics for the service performance monitoring.
	options := app.Options{
		kafkaConsumer.Configuration{
			// kafkaAuth.AuthenticationConfig{
			// 	Authentication: kafkaConsumerAuthentication,
			// 	Kerberos: kafkaAuth.KerberosConfig{
			// 		ServiceName: kafkaConsumerKerberosServiceName,
			// 		Realm:       kafkaConsumerKerberosRealm,
			// 		UseKeyTab:   kafkaConsumerKerberosUseKeytab,
			// 		Username:    kafkaConsumerKerberosUsername,
			// 		Password:    kafkaConsumerKerberosPassword,
			// 		ConfigPath:  kafkaConsumerKerberosConfigFile,
			// 		KeyTabPath:  kafkaConsumerKerberosKeytabFile,
			// 	},
			// 	TLS: tlscfg.Options{
			// 		Enabled:        kafkaConsumerTLSEnabled,
			// 		CAPath:         kafkaConsumerTLSCa,
			// 		CertPath:       kafkaConsumerTLSCert,
			// 		KeyPath:        kafkaConsumerTLSKey,
			// 		ServerName:     kafkaConsumerTLSServerName,
			// 		SkipHostVerify: kafkaConsumerTLSSkipHostVerify,
			// 	},
			// 	PlainText: kafkaAuth.PlainTextConfig{
			// 		Username:  kafkaConsumerPlaintextUsername,
			// 		Password:  kafkaConsumerPlaintextPassword,
			// 		Mechanism: kafkaConsumerPlaintextMechanism,
			// 	},
			// },
			Brokers:         strings.Split(kafkaConsumerBrokers, ","),
			Topic:           kafkaConsumerTopic,
			GroupID:         kafkaConsumerGroupID,
			ClientID:        kafkaConsumerClientID,
			ProtocolVersion: kafkaConsumerProtocolVersion,
			RackID:          kafkaConsumerRackID,
		},
		exporterParallelism,
		kafkaConsumerEncoding,
		exporterDeadlockInterval,
	}

	consumer, err := builder.CreateConsumer(logger, metrics.NullFactory, exp, options)
	if err != nil {
		logger.Fatal("Unable to create consumer", zap.Error(err))
	}
	consumer.Start()

	// All components should be terminated gracefully. For that we are listen for the SIGINT and SIGTERM signals and try
	// to gracefully shutdown the started exporter server and consumer. This ensures that established connections or
	// tasks are not interrupted.
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGTERM)

	logger.Debug("Start listining for SIGINT and SIGTERM signal")
	<-done
	logger.Info("Shutdown jaeger-exporter...")

	if err := exp.Stop(); err != nil {
		logger.Error("Graceful shutdown of the exporter server failed", zap.Error(err))
	}

	if err = consumer.Close(); err != nil {
		logger.Error("Failed to close consumer", zap.Error(err))
	}
}
