package runner

import (
	"fmt"
	"github.com/nats-io/nats.go"
)

var ncGlobal *nats.Conn

// SetNATS allows server to inject a shared connection
func SetNATS(nc *nats.Conn) { ncGlobal = nc }

func emit(topic, tenant, kind string, payload []byte) {
	if tenant == "" { tenant = "local" }
	if kind == "" { kind = "progress" }
	if ncGlobal == nil { return }
	_ = ncGlobal.Publish(fmt.Sprintf("tplug.%s.%s.%s", topic, tenant, kind), payload)
}
