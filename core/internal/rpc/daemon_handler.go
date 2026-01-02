package rpc

import (
	"context"
	"runtime"
	"time"

	"github.com/orhaniscoding/goconnect/server/internal/auth"
	"github.com/orhaniscoding/goconnect/server/internal/backend"
	"github.com/orhaniscoding/goconnect/server/internal/logger"
	"github.com/orhaniscoding/goconnect/server/internal/proto"
	"google.golang.org/protobuf/types/known/emptypb"
)

// DaemonHandler implements the DaemonServiceServer gRPC interface.
type DaemonHandler struct {
	proto.UnimplementedDaemonServiceServer
	daemon interface {
		Stop()
		GetBackend() interface {
			RequestDeviceCode(context.Context) (*backend.DeviceCodeResponse, error)
			PollDeviceToken(context.Context, string) (*backend.AuthResponse, error)
		}
		GetTokenManager() auth.TokenManager
	}
	version string
}

// NewDaemonHandler creates a new handler for the Daemon service.
func NewDaemonHandler(d interface {
	Stop()
	GetBackend() interface {
		RequestDeviceCode(context.Context) (*backend.DeviceCodeResponse, error)
		PollDeviceToken(context.Context, string) (*backend.AuthResponse, error)
	}
	GetTokenManager() auth.TokenManager
}, version string) *DaemonHandler {
	return &DaemonHandler{
		daemon:  d,
		version: version,
	}
}

func (h *DaemonHandler) GetStatus(ctx context.Context, in *proto.GetStatusRequest) (*proto.GetStatusResponse, error) {
	// For now, return a basic response. Story 2.5 covers full status.
	return &proto.GetStatusResponse{
		Status: proto.ConnectionStatus_CONNECTION_STATUS_DISCONNECTED,
	}, nil
}

func (h *DaemonHandler) GetVersion(ctx context.Context, in *emptypb.Empty) (*proto.VersionResponse, error) {
	return &proto.VersionResponse{
		Version: h.version,
		Os:      runtime.GOOS,
		Arch:    runtime.GOARCH,
	}, nil
}

func (h *DaemonHandler) Shutdown(ctx context.Context, in *emptypb.Empty) (*emptypb.Empty, error) {
	logger.Info("Shutdown requested via RPC")
	go h.daemon.Stop()
	return &emptypb.Empty{}, nil
}

func (h *DaemonHandler) Subscribe(in *proto.SubscribeRequest, stream proto.DaemonService_SubscribeServer) error {
	// Not implemented yet
	return nil
}

func (h *DaemonHandler) Login(req *proto.LoginRequest, stream proto.DaemonService_LoginServer) error {
	logger.Info("Login requested via RPC")

	ctx := stream.Context()

	// 1. Request Device Code
	codeResp, err := h.daemon.GetBackend().RequestDeviceCode(ctx)
	if err != nil {
		logger.Error("Failed to request device code", "error", err)
		return stream.Send(&proto.LoginUpdate{
			Update: &proto.LoginUpdate_Error{
				Error: &proto.LoginError{Message: "Failed to initiate login: " + err.Error()},
			},
		})
	}

	// 2. Send instructions to client
	if err := stream.Send(&proto.LoginUpdate{
		Update: &proto.LoginUpdate_Instructions{
			Instructions: &proto.LoginInstructions{
				Url:  codeResp.VerificationURI,
				Code: codeResp.UserCode,
			},
		},
	}); err != nil {
		return err
	}

	// 3. Poll for token
	ticker := time.NewTicker(time.Duration(codeResp.Interval) * time.Second)
	defer ticker.Stop()

	timeout := time.After(time.Duration(codeResp.ExpiresIn) * time.Second)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-timeout:
			return stream.Send(&proto.LoginUpdate{
				Update: &proto.LoginUpdate_Error{
					Error: &proto.LoginError{Message: "Login timed out"},
				},
			})
		case <-ticker.C:
			authResp, err := h.daemon.GetBackend().PollDeviceToken(ctx, codeResp.DeviceCode)
			if err != nil {
				if err.Error() == "authorization_pending" {
					continue
				}
				if err.Error() == "slow_down" {
					// Increase interval? For now simply continue but maybe skip a tick?
					// Standard says: increase interval by 5s.
					// We'll simplisticly just continue for now or maybe sleep a bit.
					time.Sleep(5 * time.Second)
					continue
				}

				logger.Error("Polling failed", "error", err)
				return stream.Send(&proto.LoginUpdate{
					Update: &proto.LoginUpdate_Error{
						Error: &proto.LoginError{Message: "Polling failed: " + err.Error()},
					},
				})
			}

			// Success!
			// Success!
			session := &auth.TokenSession{
				AccessToken:  authResp.AccessToken,
				RefreshToken: authResp.RefreshToken,
				Expiry:       time.Now().Add(time.Duration(authResp.ExpiresIn) * time.Second),
			}

			if err := h.daemon.GetTokenManager().SaveSession(session); err != nil {
				logger.Error("Failed to save session", "error", err)
				return stream.Send(&proto.LoginUpdate{
					Update: &proto.LoginUpdate_Error{
						Error: &proto.LoginError{Message: "Failed to save session: " + err.Error()},
					},
				})
			}
			logger.Info("Login successful and session saved", "access_token_len", len(authResp.AccessToken))

			return stream.Send(&proto.LoginUpdate{
				Update: &proto.LoginUpdate_Success{
					Success: &proto.LoginSuccess{
						// User info should come from ID token claims or userinfo endpoint
						// For now, indicate successful login without user details
						UserId: "authenticated",
						Email:  "user@authenticated",
					},
				},
			})
		}
	}
}
