package distributedlock

import (
	"testing"
)

func TestDistributedlock(t *testing.T) {
	cfg := Config{
		redisAddress:  "10.125.34.100:8008",
		redisPassword: "kr",
		timeout:       60,
	}
	lck1 := NewDistributedLock("7.0.3.1", cfg)

	var uid uint64 = 1051991376
	playInfo := "ensei_play"
	err := lck1.Lock(uid, playInfo)
	if err != SUCCESS {
		t.Errorf("acquire lock fail|%d|%s|%d", uid, playInfo, err)
	}

	err = lck1.Unlock(uid, playInfo)
	if err != SUCCESS {
		t.Errorf("release lock fail|%d|%s", uid, playInfo)
	}

	err = lck1.Lock(uid, playInfo)
	if err != SUCCESS {
		t.Errorf("acquire lock fail|%d|%s|%d", uid, playInfo, err)
	}

	lck2 := NewDistributedLock("7.0.3.2", cfg)
	err = lck2.Lock(uid, playInfo)
	if err != ERR_IN_LOCK {
		t.Errorf("acquire lock wrong|%d|%s", uid, playInfo)
	}
}
