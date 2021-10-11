package chain

import (
	"time"

	"github.com/patrickmn/go-cache"
)

var CACHE = cache.New(24*time.Hour, 1*time.Hour)
