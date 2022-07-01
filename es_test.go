package core

import (
	"github.com/leicc520/go-orm/log"
	"testing"
)

func TestEs(t *testing.T) {
	es := Log2ElasticSt{Index: "go-demo-srv", Host: "http://xxx.xx.xx.xxx:9200", User: "elastic", Password: "test@1qazse4"}
	es.Write2ES(log.ERROR, "agent", "/agent", "data=111&demo=222", "123456", "789", "454545")
}
