package core

import (
	"fmt"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"net/url"
	"testing"
)

func TestBucket(t *testing.T) {
	client, err := oss.New("oss-cn-hangzhou.aliyuncs.com",
		"LTAI4GCapv2A1QUqbSY8SsUB", "gfbXRx2n5FfvLrlzwniObpCfhTA1Ci")
	if err != nil {
		t.Error(err)
		return
	}

	bucket, err := client.Bucket("yt-caigoubu")
	if err != nil {
		t.Error(err)
		return
	}

	l, err := bucket.ListObjectsV2(oss.Prefix("images/W/WM42124+CXBY001/"))
	fmt.Println(l, err)
	fmt.Println(len(l.Objects))
	for _, item := range l.Objects {
		fmt.Printf("https://yt-caigoubu.oss-cn-hangzhou.aliyuncs.com/%s\r\n", url.QueryEscape(item.Key))
	}


}
