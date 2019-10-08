package main

import (
	"flag"
	"fmt"
	"github.com/owenliang/myf-go/cron"
	"github.com/owenliang/myf-go/cron/middlewares"
	"github.com/owenliang/myf-go/mcontext"
	"time"
)

func task1(myfCtx *mcontext.MyfContext) {
	a := 0
	fmt.Println(1 / a)
}

func task2(myfCtx *mcontext.MyfContext) {
	<- myfCtx.Done()
	fmt.Println(myfCtx.Cron)
}

func main() {
	flag.Parse()

	if myfCron, err := cron.New(); err == nil {
		myfCron.AddJob("task1111", "* * * * *", myfCron.WithMyfContext(task1))
		myfCron.AddJob("task2222", "* * * * *", middlewares.WithTimeout(5 * time.Second), myfCron.WithMyfContext(task2))
		myfCron.Run()
	} else {
		fmt.Println(err)
	}
}
