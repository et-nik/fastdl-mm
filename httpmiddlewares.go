package main

import (
	"github.com/Southclaws/swirl/memory"
	"log/slog"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/Southclaws/swirl"
)

func rateLimitMiddleware(next http.Handler, period time.Duration, limit int) http.Handler {
	ratelimiter := swirl.New(memory.New(), limit, period, period/10)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			slog.Error("Failed to split remote address", "error", err)

			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		status, allowed, err := ratelimiter.Increment(ctx, ip, 1)
		if err != nil {
			slog.Error("Failed to increment ratelimit", "error", err)

			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if !allowed {
			slog.Info("Rate limit exceeded", "ip", ip, "status", status)

			w.Header().Set("Retry-After", status.Reset.UTC().Format(time.RFC1123))

			http.Error(w, http.StatusText(http.StatusTooManyRequests), http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func ipBlockMiddleware(next http.Handler, blockList []string) http.Handler {
	blockedSubnets := make([]*net.IPNet, 0, len(blockList))
	blockedIPs := make(map[string]struct{}, len(blockList))

	for _, item := range blockList {
		if strings.Contains(item, "/") {
			_, ipNet, err := net.ParseCIDR(item)

			if err != nil {
				slog.Error("Failed to parse CIDR", "cidr", item, "error", err)

				continue
			}

			blockedSubnets = append(blockedSubnets, ipNet)
		} else if strings.Contains(item, "-") {
			parts := strings.Split(item, "-")
			if len(parts) != 2 {
				slog.Error("Invalid range", "range", item)

				continue
			}

			start := net.ParseIP(parts[0])
			end := net.ParseIP(parts[1])
			if start == nil || end == nil {
				slog.Error("Failed to parse IP range", "range", item)

				continue
			}

			ipNet := IPToCidr(start, end)
			if ipNet == nil {
				slog.Error("Failed to create CIDR from range", "range", item)

				continue
			}

			blockedSubnets = append(blockedSubnets, ipNet)
		} else {
			parsed := net.ParseIP(item)
			if parsed != nil {
				blockedIPs[item] = struct{}{}
			}
		}
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			slog.Error("Failed to split remote address", "error", err)

			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		if _, blocked := blockedIPs[ip]; blocked {
			slog.Info("Blocked IP", "ip", ip)

			http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
			return
		}

		for _, ipNet := range blockedSubnets {
			if ipNet.Contains(net.ParseIP(ip)) {
				slog.Info("Blocked IP", "ip", ip)

				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

// IPToCidr https://github.com/raspi/ip-range-to-CIDR/blob/master/lib/cidr.go
func IPToCidr(start, end net.IP) *net.IPNet {
	if !((start.To4() != nil && end.To4() != nil) || (start.To16() != nil && end.To16() != nil)) {
		return nil
	}

	if start.To4() != nil {
		start = start.To4()
	} else {
		start = start.To16()
	}

	if end.To4() != nil {
		end = end.To4()
	} else {
		end = end.To16()
	}

	mask := make([]byte, len(start))

	for idx := range start {
		mask[idx] = 255 - (start[idx] ^ end[idx])
	}

	return &net.IPNet{
		IP:   start,
		Mask: mask,
	}
}
