// Copyright 2025 Daytona Platforms Inc.
// SPDX-License-Identifier: AGPL-3.0

package proxy

import (
	"net/http"
	"net/http/httputil"
)

func (s *ProxyServer) proxyRequest(w http.ResponseWriter, r *http.Request) {
	sandboxId := r.PathValue("id")
	path := r.PathValue("path")

	target, err := s.getProxyTarget(r.Context(), sandboxId)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	proxy := &httputil.ReverseProxy{
		Transport: s.transport,
		Director: func(req *http.Request) {
			req.URL.Scheme = target.Scheme
			req.URL.Host = target.Host
			req.URL.Path = "/" + path

			if r.URL.RawQuery != "" {
				req.URL.RawQuery = r.URL.RawQuery
			}

			for k, v := range r.Header {
				req.Header[k] = v
			}

			req.Host = target.Host
		},
	}

	proxy.ServeHTTP(w, r)
}
