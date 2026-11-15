// Host API 客户端：把 sdk 的服务接口接到 TT Core 的 ConnectRPC 上。
//
// 只依赖公开的 pkg/hostv1（生成物），不 import TT 的 internal 包。

package sdk

import (
	"context"
	"errors"
	"net/http"

	"connectrpc.com/connect"

	hostv1 "github.com/luyaogod/TT/pkg/hostv1"
	"github.com/luyaogod/TT/pkg/hostv1/hostv1connect"
)

// hostClient 一组 Host API 客户端（同一 base URL + 同一实例令牌）。
type hostClient struct {
	profiles hostv1connect.ProfilesServiceClient
	conns    hostv1connect.ConnectionsServiceClient
	db       hostv1connect.DBSessionServiceClient
	ssh      hostv1connect.SSHSessionServiceClient
	events   hostv1connect.EventsServiceClient
}

func newHostClient(baseURL, token string) *hostClient {
	hc := &http.Client{}
	opts := []connect.ClientOption{
		// Connect 的默认编码是 proto；TT 的 Host API 走 JSON（与页面/RPC 同一套），
		// 这里显式选 JSON，免得两边协商成两份线上格式。
		connect.WithProtoJSON(),
		connect.WithInterceptors(connect.UnaryInterceptorFunc(
			func(next connect.UnaryFunc) connect.UnaryFunc {
				return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
					req.Header().Set("Authorization", "Bearer "+token)
					return next(ctx, req)
				}
			},
		)),
	}
	return &hostClient{
		profiles: hostv1connect.NewProfilesServiceClient(hc, baseURL, opts...),
		conns:    hostv1connect.NewConnectionsServiceClient(hc, baseURL, opts...),
		db:       hostv1connect.NewDBSessionServiceClient(hc, baseURL, opts...),
		ssh:      hostv1connect.NewSSHSessionServiceClient(hc, baseURL, opts...),
		events:   hostv1connect.NewEventsServiceClient(hc, baseURL, opts...),
	}
}

// mapErr 把 Connect 错误翻成带稳定码的 *Error（优先用 ErrorInfo 详情）。
func mapErr(err error) error {
	if err == nil {
		return nil
	}
	var ce *connect.Error
	if errors.As(err, &ce) {
		for _, d := range ce.Details() {
			msg, verr := d.Value()
			if verr != nil {
				continue
			}
			if info, ok := msg.(*hostv1.ErrorInfo); ok {
				return &Error{Code: Code(info.GetCode()), Message: info.GetMessage(), Hint: info.GetHint()}
			}
		}
		return &Error{Code: codeFromConnect(ce.Code()), Message: ce.Message()}
	}
	return &Error{Code: CodeInternal, Message: err.Error()}
}

// codeFromConnect 只在没有 ErrorInfo 详情时兜底。
func codeFromConnect(c connect.Code) Code {
	switch c {
	case connect.CodePermissionDenied:
		return CodePermissionDenied
	case connect.CodeDeadlineExceeded:
		return CodeTimeout
	case connect.CodeUnimplemented:
		return CodeUnimplemented
	case connect.CodeInvalidArgument:
		return CodeInvalidArgument
	default:
		return CodeInternal
	}
}
