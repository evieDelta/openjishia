package ratelimiter_test

import (
	"testing"
	"time"

	"github.com/eviedelta/openjishia/internal/ratelimiter"
)

func TestStandard(t *testing.T) {
	r := ratelimiter.New(ratelimiter.Config{
		Period: time.Second,
	})
	if !r.Allowed("a") {
		t.Log("key a is disallowed when it should be allowed")
		t.Fail()
	}
	if !r.Do("a") {
		t.Log("key a is disallowed when it should be allowed 2")
		t.Fail()
	}
	if r.Do("a") {
		t.Log("key a is allowed when it should be disallowed")
		t.Fail()
	}
	if !r.Do("b") {
		t.Log("key b is disallowed when it should be allowed")
		t.Fail()
	}
	if r.Do("b") {
		t.Log("key b is allowed when it should be disallowed")
	}
	time.Sleep(time.Second * 2)
	if !r.Do("a") {
		t.Log("key a is disallowed when it should be allowed 3")
	}
}

/*
func TestLeakyBucket(t *testing.T) {
	r := ratelimiter.New(ratelimiter.Config{
		Period:   time.Second * 3,
		MaxBurst: 3,
	})
	if !r.Allowed("a") {
		t.Log("key a is disallowed when it should be allowed")
		t.Fail()
	}
	if !r.Do("a") {
		t.Log("key a is disallowed when it should be allowed 2")
		t.Fail()
	}
	if !r.Do("a") {
		t.Log("key a is disallowed when it should be allowed 3")
		t.Fail()
	}
	if !r.Do("a") {
		t.Log("key a is disallowed when it should be allowed 4")
		t.Fail()
	}
	if r.Do("a") {
		t.Log("key a is allowed when it should be disallowed")
		t.Fail()
	}
	time.Sleep(time.Second)
	if !r.Do("a") {
		t.Log("key a is disallowed when it should be allowed 5")
		t.Fail()
	}
	if r.Do("a") {
		t.Log("key a is allowed when it should be disallowed 2")
		t.Fail()
	}
}
*/
