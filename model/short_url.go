package model

import (
	"context"
	"github.com/hprose/hprose-golang/v3/rpc/core"
	"net/http"
)

func ShortURLGen(id, uri string) (string, error) {

	clientContext := core.NewClientContext()
	header := make(http.Header)
	header.Set("X-Func-Name", "ShortURLGen")
	clientContext.Items().Set("httpRequestHeaders", header)
	rCtx := core.WithContext(context.Background(), clientContext)

	recs, err := grpc_t.ShortURLGen(rCtx, id, uri)
	if err != nil {
		return "", err
	}

	return recs, nil
}
