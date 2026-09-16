package cache

import "time"

type AuthCache struct {
	Keys   interface{}
	Expiry time.Time
}

func (cache *AuthCache) IsExpired() bool {

	if diff := time.Now().Sub(cache.Expiry).Hours(); diff > 0 {
		return true
	}
	return false
}
