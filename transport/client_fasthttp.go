//go:build !nethttp
// +build !nethttp

// Package transport provides streaming object-based transport over http for intra-cluster continuous
// intra-cluster communications (see README for details and usage example).
/*
 * Copyright (c) 2018-2020, NVIDIA CORPORATION. All rights reserved.
 */
package transport

import (
	"crypto/tls"
	"io"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/NVIDIA/aistore/3rdparty/glog"
	"github.com/NVIDIA/aistore/cmn"
	"github.com/valyala/fasthttp"
)

type Client interface {
	Do(req *fasthttp.Request, resp *fasthttp.Response) error
}

func whichClient() string { return "fasthttp" }

// overriding fasthttp default `const DefaultDialTimeout = 3 * time.Second`
func dialTimeout(addr string) (net.Conn, error) {
	return fasthttp.DialTimeout(addr, 10*time.Second)
}

// intra-cluster networking: fasthttp client
func NewIntraDataClient() Client {
	config := cmn.GCO.Get()

	// apply global defaults
	wbuf, rbuf := config.Net.HTTP.WriteBufferSize, config.Net.HTTP.ReadBufferSize
	if wbuf == 0 {
		wbuf = cmn.DefaultWriteBufferSize // NOTE: fasthttp uses 4K
	}
	if rbuf == 0 {
		rbuf = cmn.DefaultReadBufferSize // ditto
	}

	if !config.Net.HTTP.UseHTTPS {
		return &fasthttp.Client{
			Dial:            dialTimeout,
			ReadBufferSize:  rbuf,
			WriteBufferSize: wbuf,
		}
	}
	return &fasthttp.Client{
		Dial:            dialTimeout,
		ReadBufferSize:  rbuf,
		WriteBufferSize: wbuf,
		TLSConfig:       &tls.Config{InsecureSkipVerify: config.Net.HTTP.SkipVerify},
	}
}

func (s *streamBase) do(body io.Reader) (err error) {
	// init request & response
	req, resp := fasthttp.AcquireRequest(), fasthttp.AcquireResponse()
	req.Header.SetMethod(http.MethodPut)
	req.SetRequestURI(s.dstURL)
	req.SetBodyStream(body, -1)
	if s.streamer.compressed() {
		req.Header.Set(cmn.HdrCompress, cmn.LZ4Compression)
	}
	req.Header.Set(cmn.HdrSessID, strconv.FormatInt(s.sessID, 10))
	// do
	err = s.client.Do(req, resp)
	if err != nil {
		if verbose {
			glog.Errorf("%s: Error [%v]", s, err)
		}
		return
	}
	// handle response & cleanup
	resp.BodyWriteTo(io.Discard)
	fasthttp.ReleaseRequest(req)
	fasthttp.ReleaseResponse(resp)
	if s.streamer.compressed() {
		s.streamer.resetCompression()
	}
	return
}
