package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/encoding/protojson"
)

func TestEnumMarshalling(t *testing.T) {
	fd := &FeedData{
		Type: FeedType_FEED_TYPE_UNKNOWN,
	}
	data, err := protojson.Marshal(fd)
	assert.NoError(t, err)
	fd2 := &FeedData{
		Type: FeedType_FEED_TYPE_RED,
	}
	data2, err := protojson.Marshal(fd2)
	assert.NoError(t, err)
	assert.Equal(t, `{}`, string(data))
	assert.Equal(t, `{"type":"FEED_TYPE_RED"}`, string(data2))
}
