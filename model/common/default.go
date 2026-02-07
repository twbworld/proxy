package common

import (
	"github.com/twbworld/proxy/model/config"
)

type ClashVlessBase struct {
	*config.Proxies
	RealityOpts bool `json:"reality-opts,omitempty"`
	GrpcOpts    bool `json:"grpc-opts,omitempty"`
	XhttpOpts   bool `json:"xhttp-opts,omitempty"`
}

// 针对 TCP (Reality) 的结构
type ClashVlessReality struct {
	*config.Proxies
	GrpcOpts  bool `json:"grpc-opts,omitempty"`
	XhttpOpts bool `json:"xhttp-opts,omitempty"`
}

// 针对 gRPC 的结构
type ClashVlessGrpc struct {
	*config.Proxies
	RealityOpts bool `json:"reality-opts,omitempty"`
	XhttpOpts   bool `json:"xhttp-opts,omitempty"`
}
