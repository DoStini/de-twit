package common

import (
	"context"
	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multicodec"
	"github.com/multiformats/go-multihash"
	"log"
)

func GenerateCid(ctx context.Context, key string) (cid.Cid, error) {
	logger := ctx.Value("logger").(*log.Logger)

	pref := cid.Prefix{
		Version:  1,
		Codec:    uint64(multicodec.Raw),
		MhType:   multihash.SHA2_256,
		MhLength: -1, // default length
	}
	c, err := pref.Sum([]byte(key))
	if err != nil {
		logger.Fatalf(err.Error())
	}
	return c, err
}

func Contains(list []cid.Cid, val cid.Cid) bool {
	for _, b := range list {
		if b == val {
			return true
		}
	}
	return false
}

func FindIndex(list []cid.Cid, val cid.Cid) int {
	for i, b := range list {
		if b == val {
			return i
		}
	}
	return -1
}

func RemoveIndex(list []cid.Cid, index int) []cid.Cid {
	last := len(list) - 1
	list[index] = list[last]
	list = list[:last]

	return list
}
