package common

import (
	"context"
	"github.com/ipfs/go-cid"
	"github.com/multiformats/go-multicodec"
	"github.com/multiformats/go-multihash"
	"log"
	"os"
)

func GetLogger(ctx context.Context) *log.Logger {
	var logger *log.Logger
	logger = ctx.Value("logger").(*log.Logger)
	if logger == nil {
		logger = log.New(os.Stdin, "", log.Ltime|log.Lshortfile)
	}

	return logger
}

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

func Contains[K comparable](list []K, val K) bool {
	for _, b := range list {
		if b == val {
			return true
		}
	}
	return false
}

func FindIndex[K comparable](list []K, val K) int {
	for i, b := range list {
		if b == val {
			return i
		}
	}
	return -1
}

func RemoveIndex[K any](list []K, index int) []K {
	last := len(list) - 1
	list[index] = list[last]
	list = list[:last]

	return list
}
