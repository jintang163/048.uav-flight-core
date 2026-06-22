package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	Server          ServerConfig     `mapstructure:"server"`
	Database        DatabaseConfig   `mapstructure:"database"`
	Redis           RedisConfig      `mapstructure:"redis"`
	NSQ             NSQConfig        `mapstructure:"nsq"`
	JWT             JWTConfig        `mapstructure:"jwt"`
	MAVLink         MAVLinkConfig    `mapstructure:"mavlink"`
	WebSocket       WebSocketConfig  `mapstructure:"websocket"`
	MinIO           MinIOConfig      `mapstructure:"minio"`
	Alert           AlertConfig      `mapstructure:"alert"`
	WebRTC          WebRTCConfig     `mapstructure:"webrtc"`
	YOLOv8API       string           `mapstructure:"yolov8_api"`
	YOLOv8APIKey    string           `mapstructure:"yolov8_api_key"`
	ExternalServices map[string]string `mapstructure:"external_services"`
}

type ServerConfig struct {
	Port int    `mapstructure:"port"`
	Mode string `mapstructure:"mode"`
}

type DatabaseConfig struct {
	Host        string `mapstructure:"host"`
	Port        int    `mapstructure:"port"`
	User        string `mapstructure:"user"`
	Password    string `mapstructure:"password"`
	DBName      string `mapstructure:"dbname"`
	Charset     string `mapstructure:"charset"`
	MaxIdleConns int   `mapstructure:"max_idle_conns"`
	MaxOpenConns int   `mapstructure:"max_open_conns"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

type NSQConfig struct {
	ProducerAddr      string `mapstructure:"producer_addr"`
	ConsumerLookupAddr string `mapstructure:"consumer_lookup_addr"`
}

type JWTConfig struct {
	Secret           string `mapstructure:"secret"`
	AccessTokenExpire  int  `mapstructure:"access_token_expire"`
	RefreshTokenExpire int  `mapstructure:"refresh_token_expire"`
}

type MAVLinkConfig struct {
	TCPPort          int `mapstructure:"tcp_port"`
	UDPPort          int `mapstructure:"udp_port"`
	HeartbeatInterval int `mapstructure:"heartbeat_interval"`
	Timeout          int `mapstructure:"timeout"`
}

type WebSocketConfig struct {
	TelemetryInterval int `mapstructure:"telemetry_interval"`
	MaxConnections   int `mapstructure:"max_connections"`
}

type MinIOConfig struct {
	Endpoint        string `mapstructure:"endpoint"`
	AccessKey       string `mapstructure:"access_key"`
	SecretKey       string `mapstructure:"secret_key"`
	BucketLogs      string `mapstructure:"bucket_logs"`
	BucketMissions  string `mapstructure:"bucket_missions"`
	UseSSL          bool   `mapstructure:"use_ssl"`
}

type AlertConfig struct {
	SMSAPI          string `mapstructure:"sms_api"`
	SMSAPIKey       string `mapstructure:"sms_api_key"`
	EmailSMTPHost   string `mapstructure:"email_smtp_host"`
	EmailSMTPPort   int    `mapstructure:"email_smtp_port"`
	EmailUser       string `mapstructure:"email_user"`
	EmailPassword   string `mapstructure:"email_password"`
	LowBatteryThreshold float64 `mapstructure:"low_battery_threshold"`
	SignalLossThreshold int `mapstructure:"signal_loss_threshold"`
}

type WebRTCConfig struct {
	MediaServerHost string `mapstructure:"media_server_host"`
	MediaServerPort int    `mapstructure:"media_server_port"`
	WHIPEndpoint    string `mapstructure:"whip_endpoint"`
	WHEPEndpoint    string `mapstructure:"whep_endpoint"`
}

var AppConfig *Config

func LoadConfig(path string) (*Config, error) {
	viper.SetConfigFile(path)
	viper.SetConfigType("yaml")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	AppConfig = &config
	return &config, nil
}

func (c *DatabaseConfig) DSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=True&loc=Local",
		c.User, c.Password, c.Host, c.Port, c.DBName, c.Charset)
}

func (c *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}
