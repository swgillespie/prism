package config

type IngestEventListener struct {
	SQSEndpoint                 string `yaml:"sqs_endpoint"`
	SQSQueueEndpoint            string `yaml:"sqs_queue_endpoint"`
	DestinationBucket           string `yaml:"destination_bucket"`
	MaxMessages                 int    `yaml:"max_messages"`
	MessageHandleTimeoutSeconds int    `yaml:"message_handle_timeout_seconds"`
	PollTimeoutSeconds          int    `yaml:"poll_timeout_seconds"`
}
