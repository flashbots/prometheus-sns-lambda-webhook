package config

type Config struct {
	Log     Log
	Webhook Webhook
}

type Log struct {
	Level string
	Mode  string
}

type Webhook struct {
	Method      string
	IncludeBody bool
	Url         string
}
