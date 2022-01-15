package test

import (
	"log"
	"time"
)

type pollIntervalBuilder interface{
	PollEvery(d time.Duration) untilTrueBuilder
}

type untilTrueBuilder interface{
	Until(condition conditionFunc) awaitBuilder
}

type awaitBuilder interface{
	Start() 
}

type conditionFunc func() bool
type await struct{
	timeout time.Duration
	pollInterval time.Duration
	condition conditionFunc
}

func Await(d time.Duration) pollIntervalBuilder{
	return await{
		timeout: d,
	}
}

func (a await) PollEvery(d time.Duration) untilTrueBuilder{
	a.pollInterval = d
	return a
}

func (a await) Until(condition conditionFunc) awaitBuilder{
	a.condition = condition
	return a
}

func (a await) Start(){
	ticker := time.NewTicker(a.pollInterval)
	defer ticker.Stop()

    timeout := make(chan bool)
	go func(){
		time.Sleep(a.timeout)
		timeout <- true
	}()

    go func() {
        for {
            select {
            case <-timeout:
				log.Print("Timeout without condition pass")
                return
            case <-ticker.C:
                if a.condition(){
					log.Print("Condition passed")
					timeout <- true
				}
            }
        }
    }()
	
	<- timeout
}

