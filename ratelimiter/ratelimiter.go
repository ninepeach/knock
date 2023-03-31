package ratelimiter

import (
	"encoding/json"
	"io/ioutil"
	"net"
	"os"
	"sync"
	"time"
)

type IPAccess struct {
	AccessCount int
	LastAccess  time.Time
}

type RateLimiter struct {
	MaxAccesses   int
	AccessTimeout time.Duration
	AccessMap     map[string]*IPAccess
	Blacklist     map[string]bool
	mu            sync.Mutex
	blacklistFile string
}

func NewRateLimiter(maxAccesses int, accessTimeout time.Duration, blacklistFile string) (*RateLimiter, error) {
	accessMap := make(map[string]*IPAccess)
	blacklist := make(map[string]bool)

	// Load blacklist from disk
	if len(blacklistFile) > 0 {
		if _, err := os.Stat(blacklistFile); err == nil {
			blacklistData, err := ioutil.ReadFile(blacklistFile)
			if err != nil {
				return nil, err
			}
			if err := json.Unmarshal(blacklistData, &blacklist); err != nil {
				return nil, err
			}
		}
	}

	return &RateLimiter{
		MaxAccesses:   maxAccesses,
		AccessTimeout: accessTimeout,
		AccessMap:     accessMap,
		Blacklist:     blacklist,
		blacklistFile: blacklistFile,
	}, nil
}

func (rl *RateLimiter) IsAllowed(ip net.IP) bool {
	ipString := ip.String()
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// Check if IP is in blacklist
	if rl.Blacklist[ipString] {
		return false
	}

	// Check if IP has exceeded maximum access count
	if access, ok := rl.AccessMap[ipString]; ok {
		if time.Since(access.LastAccess) < rl.AccessTimeout {
			if access.AccessCount >= rl.MaxAccesses {
				rl.Blacklist[ipString] = true
				rl.saveBlacklist()
				return false
			} else {
				access.AccessCount++
				access.LastAccess = time.Now()
			}
		} else {
			access.AccessCount = 1
			access.LastAccess = time.Now()
		}
	} else {
		rl.AccessMap[ipString] = &IPAccess{AccessCount: 1, LastAccess: time.Now()}
	}

	return true
}

func (rl *RateLimiter) saveBlacklist() error {
	blacklistData, err := json.Marshal(rl.Blacklist)
	if err != nil {
		return err
	}
	if err := ioutil.WriteFile(rl.blacklistFile, blacklistData, 0644); err != nil {
		return err
	}
	return nil
}
