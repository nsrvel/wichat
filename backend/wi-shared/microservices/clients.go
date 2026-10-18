package microservices

import (
	"github.com/wichat/wichat/backend/wi-shared/infra/logger"
	msUser "github.com/wichat/wichat/backend/wi-shared/microservices/services/ms-user"
	"github.com/wichat/wichat/backend/wi-shared/microservices/utils"
	msgrpc "github.com/wichat/wichat/backend/wi-shared/microservices/utils/grpc"
	"google.golang.org/grpc"
)

// Config wires outbound microservice clients (timeouts, retries, dial targets).
type Config struct {
	Log      logger.Logger
	Settings utils.Settings
	UserAddr string
}

// Clients holds typed outbound facades. Open once per process; inject into handlers.
type Clients struct {
	User msUser.Service

	userConn *grpc.ClientConn
}

// Open dials configured backends and returns Clients. Caller must Close on shutdown.
func Open(cfg Config) (*Clients, error) {
	conn, err := msgrpc.Dial(cfg.UserAddr, msgrpc.DefaultDialOptions)
	if err != nil {
		return nil, err
	}

	return &Clients{
		User: msUser.Service{
			GRPC: msUser.NewGRPC(conn, cfg.Settings, cfg.Log),
		},
		userConn: conn,
	}, nil
}

// Close releases outbound connections.
func (c *Clients) Close() error {
	if c == nil || c.userConn == nil {
		return nil
	}

	return c.userConn.Close()
}
