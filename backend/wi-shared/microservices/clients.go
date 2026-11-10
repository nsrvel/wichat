package microservices

import (
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	msChat "github.com/wichat/wichat/backend/wi-shared/microservices/services/ms-chat"
	msNotification "github.com/wichat/wichat/backend/wi-shared/microservices/services/ms-notification"
	msUser "github.com/wichat/wichat/backend/wi-shared/microservices/services/ms-user"
	"github.com/wichat/wichat/backend/wi-shared/microservices/utils"
	msgrpc "github.com/wichat/wichat/backend/wi-shared/microservices/utils/grpc"
	"google.golang.org/grpc"
)

// Config wires outbound microservice clients (timeouts, retries, dial targets).
type Config struct {
	Log              logger.Logger
	Settings         utils.Settings
	UserAddr         string
	ChatAddr         string
	NotificationAddr string
}

// Clients holds typed outbound facades. Open once per process; inject into handlers.
type Clients struct {
	User         msUser.Service
	Chat         msChat.Service
	Notification msNotification.Service

	userConn         *grpc.ClientConn
	chatConn         *grpc.ClientConn
	notificationConn *grpc.ClientConn
}

// Open dials configured backends and returns Clients. Caller must Close on shutdown.
func Open(cfg Config) (*Clients, error) {
	conn, err := msgrpc.Dial(cfg.UserAddr, msgrpc.DefaultDialOptions)
	if err != nil {
		return nil, err
	}

	clients := &Clients{
		User: msUser.Service{
			GRPC: msUser.NewGRPC(conn, cfg.Settings, cfg.Log),
		},
		userConn: conn,
	}

	if cfg.ChatAddr != "" {
		chatConn, err := msgrpc.Dial(cfg.ChatAddr, msgrpc.DefaultDialOptions)
		if err != nil {
			_ = conn.Close()
			return nil, err
		}
		clients.Chat = msChat.Service{
			GRPC: msChat.NewGRPC(chatConn, cfg.Settings, cfg.Log),
		}
		clients.chatConn = chatConn
	}

	if cfg.NotificationAddr != "" {
		notifConn, err := msgrpc.Dial(cfg.NotificationAddr, msgrpc.DefaultDialOptions)
		if err != nil {
			_ = clients.closeUserAndChat()
			return nil, err
		}
		clients.Notification = msNotification.Service{
			GRPC: msNotification.NewGRPC(notifConn, cfg.Settings, cfg.Log),
		}
		clients.notificationConn = notifConn
	}

	return clients, nil
}

// Close releases outbound connections.
func (c *Clients) Close() error {
	if c == nil {
		return nil
	}
	var first error
	if c.notificationConn != nil {
		if err := c.notificationConn.Close(); err != nil && first == nil {
			first = err
		}
	}
	if c.chatConn != nil {
		if err := c.chatConn.Close(); err != nil && first == nil {
			first = err
		}
	}
	if c.userConn != nil {
		if err := c.userConn.Close(); err != nil && first == nil {
			first = err
		}
	}

	return first
}

func (c *Clients) closeUserAndChat() error {
	var first error
	if c.chatConn != nil {
		if err := c.chatConn.Close(); err != nil {
			first = err
		}
	}
	if c.userConn != nil {
		if err := c.userConn.Close(); err != nil && first == nil {
			first = err
		}
	}

	return first
}
