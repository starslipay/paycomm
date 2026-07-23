package xerror

import (
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
