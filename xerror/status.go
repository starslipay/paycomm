package xerror

import (
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/anypb"
)

// NewBizStatus 构造携带业务错误的grpc status
// grpcCode：框架层级错误码，统一建议 codes.Internal
// bizCode：业务自定义错误码
func NewBizStatus(grpcCode codes.Code, bizCode int64, msg string) *status.Status {
	bizErr := &BizError{
		Code:    bizCode,
		Message: msg,
	}
	anyData, _ := anypb.New(bizErr)
	st := status.New(grpcCode, msg)
	st, _ = st.WithDetails(anyData)
	return st
}

// NewBizError 返回error，直接在rpc handler return
func NewBizError(grpcCode codes.Code, bizCode int64, msg string) error {
	return NewBizStatus(grpcCode, bizCode, msg).Err()
}

// ParseBizError 从error反向解析出自定义BizError
func ParseBizError(err error) (*BizError, bool) {
	st, ok := status.FromError(err)
	if !ok {
		return nil, false
	}
	for _, d := range st.Details() {
		if anyData, ok := d.(*anypb.Any); ok {
			var bizErr BizError
			if err := anyData.UnmarshalTo(&bizErr); err == nil {
				return &bizErr, true
			}
		}
	}
	return nil, false
}

// HandleRPCError 处理RPC错误码
// err: rpc调用返回的错误
// serviceName: 服务名，用于非业务错误场景的错误信息
// 返回处理后的error，包含业务错误码或服务调用错误信息
func HandleRPCError(err error, serviceName string) error {
	if err == nil {
		return nil
	}

	st, ok := status.FromError(err)
	if !ok {
		return fmt.Errorf("%s 服务调用错误: %w", serviceName, err)
	}

	switch st.Code() {
	case codes.Internal:
		if bizErr, ok := ParseBizError(err); ok {
			return NewBizError(codes.Internal, bizErr.Code, bizErr.Message)
		}
		return NewBizError(codes.Internal, 1000, fmt.Sprintf("%s 服务内部错误", serviceName))
	case codes.Unavailable:
		return NewBizError(codes.Internal, 1001, fmt.Sprintf("%s 服务不可达", serviceName))
	case codes.DeadlineExceeded:
		return NewBizError(codes.Internal, 1002, fmt.Sprintf("%s 调用超时", serviceName))
	case codes.Canceled:
		return NewBizError(codes.Internal, 1003, fmt.Sprintf("%s 上下文已取消", serviceName))
	default:
		return NewBizError(codes.Internal, 1004, fmt.Sprintf("%s 服务调用失败: %s", serviceName, st.Message()))
	}
}
