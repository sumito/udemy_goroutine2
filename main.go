package main

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"
)

var data []int = []int{1, 2, 3, 4, 5}

func writeToChan(writeChan chan<- int) {
	defer close(writeChan)
	for i := range data {
		writeChan <- i
	}
}
func chanOwner() <-chan int {
	results := make(chan int, 5)
	go func() {
		defer close(results)
		for i := 1; i <= 5; i++ {
			results <- i
		}
	}()

	return results
}
func consumer(results <-chan int) {
	for result := range results {
		fmt.Println(result)
	}
}

type Result struct {
	Response *os.File
	Error    error
}

func CheckFiles(done <-chan interface{}, filenames ...string) <-chan Result {
	results := make(chan Result)

	go func() {
		defer close(results)
		for _, filename := range filenames {
			var result Result
			file, err := os.Open(filename)
			result = Result{file, err}

			if err != nil {
				fmt.Println(err)
				return
			}
			select {
			case <-done:
				return
			case results <- result:
			}
		}
	}()

	return results
}

/*
//194
func DoSumething(done <-chan interface{}, strings <-chan string) <-chan interface{} {
	completed := make(chan interface{})

	go func() {
		defer fmt.Println("DoSumething Done")
		defer close(completed)

		for {
			select {
			case s := <-strings:
				fmt.Println(s)
			case <-done:
				return
			}
		}

	}()
	return completed
}
*/

func DoSumething(done chan interface{}) <-chan int {

	readStream := make(chan int)

	go func() {
		defer fmt.Println("DoSomething done")
		defer close(readStream)

		for {
			select {
			case readStream <- rand.Intn(100):
			case <-done:
				return
			}
		}
	}()

	return readStream
}

func or(channels ...<-chan interface{}) <-chan interface{} {
	switch len(channels) {
	case 0:
		return nil
	case 1:
		return channels[0]
	}
	orDone := make(chan interface{})

	go func() {
		defer close(orDone)

		switch len(channels) {
		case 2:
			select {
			case <-channels[0]:
			case <-channels[1]:
			}
		default:
			{
				select {
				case <-channels[0]:
				case <-channels[1]:
				case <-channels[2]:
				case <-or(append(channels[3:], orDone)...):
				}
			}
		}
	}()

	return orDone
}
func signal(after time.Duration) <-chan interface{} {
	done := make(chan interface{})
	go func() {
		defer close(done)
		time.Sleep(after)
	}()
	return done
}

/*
func double(sl []int) []int {

		doubleSlice := make([]int, len(sl))
		for i, v := range sl {
			doubleSlice[i] = v * 2
		}
		return doubleSlice
	}

	func add(sl []int) []int {
		addSlice := make([]int, len(sl))
		for i, v := range sl {
			addSlice[i] = v + 1
		}
		return addSlice
	}
*/
func double(done <-chan interface{}, intStream <-chan int) <-chan int {
	doubleStream := make(chan int)
	go func() {
		defer close(doubleStream)
		for i := range intStream {
			select {
			case <-done:
				return
			case doubleStream <- i * 2:
			}
		}
	}()
	return doubleStream
}
func add(done <-chan interface{}, intStream <-chan int) <-chan int {
	addStream := make(chan int)
	go func() {
		defer close(addStream)
		for i := range intStream {
			select {
			case <-done:
				return
			case addStream <- i + 1:
			}
		}
	}()
	return addStream
}

func generator(done <-chan interface{}, integers ...int) <-chan int {
	intStream := make(chan int)
	go func() {
		defer close(intStream)

		for _, i := range integers {
			select {
			case <-done:
				return
			case intStream <- i:
			}
		}
	}()

	return intStream
}

func repeatFunc(done <-chan interface{}, fn func() interface{}) <-chan interface{} {
	valueStream := make(chan interface{})

	go func() {
		defer close(valueStream)
		for {
			select {
			case <-done:
				return
			case valueStream <- fn():
			}
		}
	}()

	return valueStream
}
func take(done <-chan interface{}, valueStream <-chan interface{}, num int) <-chan interface{} {
	takeStream := make(chan interface{})

	go func() {
		defer close(takeStream)
		for i := 0; i < num; i++ {
			select {
			case <-done:
				return
			case takeStream <- <-valueStream:
			}
		}
	}()

	return takeStream
}

func toInt(done <-chan interface{}, valueStream <-chan interface{}) <-chan int {
	intStream := make(chan int)

	go func() {
		defer close(intStream)

		for v := range valueStream {
			select {
			case <-done:
				return
			case intStream <- v.(int):
			}
		}
	}()

	return intStream
}

func primeFinder(done <-chan interface{}, intStream <-chan int) <-chan interface{} {
	primeStream := make(chan interface{})
	go func() {
		defer close(primeStream)

	L:
		for i := range intStream {
			for div := 2; div < i; div++ {
				if i%div == 0 {
					continue L
				}
			}

			select {
			case <-done:
				return
			case primeStream <- i:
			}
		}
	}()

	return primeStream
}

func random() interface{} {
	return rand.Intn(500000000)
}

func fanIn(done <-chan interface{}, channels ...<-chan interface{}) <-chan interface{} {
	var wg sync.WaitGroup

	multiplexedStream := make(chan interface{})

	multiplex := func(c <-chan interface{}) {
		defer wg.Done()
		for i := range c {
			select {
			case <-done:
				return
			case multiplexedStream <- i:
			}
		}
	}

	for _, c := range channels {
		wg.Add(1)
		go multiplex(c)
	}

	go func() {
		wg.Wait()
		close(multiplexedStream)
	}()

	return multiplexedStream
}

func generatorForOrDone() <-chan int {
	intStream := make(chan int)
	go func() {
		defer close(intStream)
		for i := 0; i <= 100; i++ {
			intStream <- i
		}
	}()
	return intStream
}
func signalForOrDone(after time.Duration) <-chan interface{} {
	done := make(chan interface{})

	go func() {
		defer close(done)
		defer fmt.Println("signal Done")
		time.Sleep(after)
	}()

	return done
}

func orDone(done <-chan interface{}, c <-chan int) <-chan interface{} {
	valCh := make(chan interface{})
	go func() {
		defer close(valCh)
		for {
			select {
			case <-done:
				return
			case v, ok := <-c:
				if !ok {
					return
				}
				select {
				case valCh <- v:
				case <-done:
				}
			}
		}
	}()
	return valCh
}

func generatorTee(done <-chan interface{}) <-chan interface{} {
	intChan := make(chan interface{})

	go func() {
		defer fmt.Println("generator Tee Done")
		defer close(intChan)

		for i := 0; i < 100; i++ {
			select {
			case <-done:
				return
			case intChan <- i:
			}
		}
	}()

	return intChan
}
func orDoneTee(done, c <-chan interface{}) <-chan interface{} {
	valChan := make(chan interface{})

	go func() {
		defer close(valChan)
		for {
			select {
			case <-done:
			case v, ok := <-c:
				if !ok {
					return
				}

				select {
				case valChan <- v:
				case <-done:
				}
			}
		}
	}()

	return valChan
}

func tee(done, c <-chan interface{}) (<-chan interface{}, <-chan interface{}) {
	out1 := make(chan interface{})
	out2 := make(chan interface{})

	go func() {
		defer close(out1)
		defer close(out2)

		for v := range orDoneTee(done, c) {
			var Out1, Out2 = out1, out2
			for i := 0; i < 2; i++ {
				select {
				case Out1 <- v:
					Out1 = nil
				case Out2 <- v:
					Out2 = nil
				}
			}
		}
	}()

	return out1, out2

}

func generateVals() <-chan <-chan interface{} {
	chanStream := make(chan (<-chan interface{}))

	go func() {
		defer close(chanStream)
		for i := 0; i < 10; i++ {
			stream := make(chan interface{}, 1)
			stream <- i
			close(stream)
			chanStream <- stream
		}
	}()

	return chanStream
}
func bridge(done <-chan interface{}, chanCh <-chan <-chan interface{}) <-chan interface{} {
	valCh := make(chan interface{})

	go func() {
		defer close(valCh)
		for {
			var ch <-chan interface{}
			select {
			case maybeCh, ok := <-chanCh:
				if !ok {
					return
				}
				ch = maybeCh
			case <-done:
				return
			}

			for v := range orDoneBridge(done, ch) {
				select {
				case valCh <- v:
				case <-done:

				}
			}
		}
	}()

	return valCh
}

func orDoneBridge(done <-chan interface{}, c <-chan interface{}) <-chan interface{} {
	valCh := make(chan interface{})
	go func() {
		defer close(valCh)
		for {
			select {
			case <-done:
				return
			case v, ok := <-c:
				if !ok {
					return
				}
				select {
				case valCh <- v:
				case <-done:
				}
			}
		}
	}()
	return valCh
}

func shortProcess(done <-chan interface{}) (bool, error) {
	shortWork := time.NewTicker(1 * time.Second)

	select {
	case <-done:
		return false, fmt.Errorf("cancel")
	case <-shortWork.C:
	}
	return true, nil
}
func longProcess(done <-chan interface{}) (bool, error) {
	longWork := time.NewTicker(5 * time.Second)

	select {
	case <-done:
		return false, fmt.Errorf("cancel")
	case <-longWork.C:
	}

	return true, nil
}

func shortProcessContext(ctx context.Context) (bool, error) {
	shortWork := time.NewTicker(1 * time.Second)

	ctx, cancel := context.WithTimeout(ctx, 1*time.Millisecond)
	defer cancel()

	select {
	case <-ctx.Done():
		return false, fmt.Errorf("cancel")
	case <-shortWork.C:
	}
	return true, nil
}

func longProcessContext(ctx context.Context) (bool, error) {
	longWork := time.NewTicker(5 * time.Second)

	select {
	case <-ctx.Done():
		return false, fmt.Errorf("cancel")
	case <-longWork.C:
	}

	return true, nil
}

func shortProcessDeadline(ctx context.Context) (bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()

	switch isDone, err := doSomething(ctx); {
	case err != nil:
		return false, err
	case isDone:
		return isDone, nil
	}
	return false, fmt.Errorf("unsupported")
}
func longProcessDeadline(ctx context.Context) (bool, error) {
	longWork := time.NewTicker(5 * time.Second)

	select {
	case <-ctx.Done():
		return false, ctx.Err()
	case <-longWork.C:
	}

	return true, nil
}
func doSomething(ctx context.Context) (bool, error) {
	if deadline, ok := ctx.Deadline(); ok {
		if deadline.Sub(time.Now().Add(2*time.Second)) <= 0 {
			return false, context.DeadlineExceeded
		}
	}
	select {
	case <-ctx.Done():
		return false, ctx.Err()
	case <-time.After(3 * time.Second):
	}
	return true, nil
}

type ctxKey int

const (
	ctxUserID ctxKey = iota
	ctxAuthToken
)

func Set(userID, authToken string) context.Context {
	ctx := context.WithValue(context.Background(), ctxUserID, userID)
	ctx = context.WithValue(ctx, ctxAuthToken, authToken)
	return ctx
}
func Get(ctx context.Context) {
	fmt.Printf("userID: %v,authToken: %v\n", ctx.Value(ctxUserID).(string), ctx.Value(ctxAuthToken).(string))
}

func worker(workerNum int, ch <-chan int, wg *sync.WaitGroup) {
	for num := range ch {
		time.Sleep(1 * time.Second)
		fmt.Printf("worker %v: %v\n", workerNum, num)
		wg.Done()
	}

}

func readGoFile(path string) chan string {
	promise := make(chan string)

	go func() {
		defer close(promise)
		content, err := os.ReadFile(path)
		if err != nil {
			fmt.Printf("read error %v\n", err.Error())
		} else {
			promise <- string(content)
		}
	}()

	return promise
}

func printFunc(futureSource chan string) chan []string {
	promise := make(chan []string)

	go func() {
		defer close(promise)
		var result []string
		for _, line := range strings.Split(<-futureSource, "\n") {
			if strings.HasPrefix(line, "func ") {
				result = append(result, line)
			}
		}
		promise <- result
	}()

	return promise
}

func DoWorkVer1(done <-chan interface{}, pulseInterval time.Duration) (<-chan interface{}, <-chan time.Time) {
	heartbeat := make(chan interface{})
	results := make(chan time.Time)

	go func() {
		defer close(heartbeat)
		defer close(results)

		heartbeatPulse := time.Tick(pulseInterval)
		workGen := time.Tick(2 * pulseInterval)

		sendHeartBeatPulse := func() {
			select {
			case heartbeat <- struct{}{}:
			default:
			}
		}
		sendResult := func(result time.Time) {
			for {
				select {
				case <-done:
					return
				case <-heartbeatPulse:
					sendHeartBeatPulse()
				case results <- result.Local():
				}
			}
		}

		for {
			//for i := 0; i < 2; i++ {
			select {
			case <-done:
				return
			case <-heartbeatPulse:
				sendHeartBeatPulse()

			case result := <-workGen:
				sendResult(result)
			}
		}
	}()
	return heartbeat, results
}

func DoWorkVer2(done <-chan interface{}) (<-chan interface{}, <-chan int) {
	heartbeatStream := make(chan interface{}, 1)
	workStream := make(chan int)
	go func() {
		defer close(heartbeatStream)
		defer close(workStream)

		for i := 0; i < 10; i++ {
			select {
			case heartbeatStream <- struct{}{}:
			default:

			}
			select {
			case <-done:
				return
			case workStream <- rand.Intn(10):
			}
		}
	}()

	return heartbeatStream, workStream
}

type startGoroutinFn func(done <-chan interface{}, pulseInterval time.Duration) (heatbeat <-chan interface{})

func DoWorkFn(done <-chan interface{}, intList ...int) (startGoroutinFn, <-chan interface{}) {
	intStream := make(chan interface{})

	doWork := func(done <-chan interface{}, pulseInterval time.Duration) <-chan interface{} {
		heartbeat := make(chan interface{})

		go func() {
			pulse := time.Tick(pulseInterval)
		valueLoop:
			for _, i := range intList {
				if i < 0 {
					log.Printf("negative value: %v\n", i)
				}
				for {
					select {
					case <-pulse:
						select {
						case heartbeat <- struct{}{}:
						default:

						}
					case intStream <- i:
						continue valueLoop
					case <-done:
						return
					}
				}
			}
		}()
		return heartbeat
	}
	return doWork, intStream
}

func newSteward(timeout time.Duration, startGroutine startGoroutinFn) startGoroutinFn {
	return func(done <-chan interface{}, pulseInterval time.Duration) <-chan interface{} {
		heartbeat := make(chan interface{})

		go func() {
			defer close(heartbeat)

			var wardOone chan interface{}
			var wardHeartbeat <-chan interface{}

			startWard := func() {
				wardOone = make(chan interface{})
				wardHeartbeat = startGroutine(wardOone, timeout/2)

			}
			startWard()
			pulse := time.Tick(pulseInterval)

		monitorLoop:
			for {
				timeoutSignal := time.After(timeout)

				for {
					select {
					case <-pulse:
						select {
						case heartbeat <- struct{}{}:
						default:
						}
					case <-wardHeartbeat:
						continue monitorLoop
					case <-timeoutSignal:
						log.Println("steward: ward is dead;restaring")
						close(wardOone)
						startWard()
					case <-done:
						return
					}
				}
			}
		}()

		return heartbeat
	}
}

func orHeartbeat(channels ...chan interface{}) <-chan interface{} {
	switch len(channels) {
	case 0:
		return nil
	case 1:
		return channels[0]
	}
	orDone := make(chan interface{})

	go func() {

		defer close(orDone)

		switch len(channels) {
		case 2:
			select {
			case <-channels[0]:
			case <-channels[1]:
			}
		default:
			select {
			case <-channels[0]:
			case <-channels[1]:
			case <-channels[2]:
				//case <-or(append(channels[3:], orDone)...):
			}
		}
	}()

	return orDone

}

func main() {

	//192 拘束 codebase rule
	/*
		handleData := make(chan int)
		go writeToChan(handleData)
		for integer := range handleData {
			fmt.Println(integer)
		}
	*/
	//192 拘束-2
	/*
		results := chanOwner()
		consumer(results)
	*/

	//193 エラーハンドリング
	/*
		done := make(chan interface{})
		defer close(done)
		filenames := []string{"main.go", "x.go"}
		for result := range CheckFiles(done, filenames...) {
			if result.Error != nil {
				log.Println("error: %v", result.Error)
				continue
			}
			fmt.Println("Response: %v\n", result.Response.Name())
		}
	*/

	//194_
	/*
		done := make(chan interface{})
		completed := DoSumething(done, nil)
		go func() {
			time.Sleep(2 * time.Second)
			close(done)
		}()

		<-completed

		fmt.Println("Main done")
	*/

	//195
	/*
		done := make(chan interface{})
		readStream := DoSumething(done)

		for i := 1; i <= 3; i++ {
			fmt.Println(<-readStream)
		}
		close(done)

		time.Sleep(1 * time.Second)

		fmt.Println("Main done")
	*/

	//196 複数の終了を伝えるチャネルをまとめて１つでも閉じらられたら終了を出力
	/*
		start := time.Now()
		<-or(signal(time.Hour), signal(time.Minute), signal(time.Second))
		fmt.Printf("done after: %v\n", time.Since(start))
	*/

	//198 パイプライン データを受け取って、何らかの処理を行ってまたどこかに流す
	/*
		ints := []int{1, 2, 3, 4, 5}
		for _, v := range double(add(ints)) {
			fmt.Println(v)
		}
	*/

	//198 パイプライン
	/*
		done := make(chan interface{})
		defer close(done)
		intStream := generator(done, 1, 2, 3, 4, 5)
		for v := range double(done, add(done, intStream)) {
			fmt.Println(v)
		}
	*/

	//200 ファンアウト、ファンイン
	/*
		done := make(chan interface{})
		defer close(done)

		randIntStream := toInt(done, repeatFunc(done, random))

		start := time.Now()

		// normal process
		//	for prime := range take(done, primeFinder(done, randIntStream), 10) {
		//		fmt.Println(prime)
		//	}

		//funout
		numFinders := runtime.NumCPU()
		fmt.Printf("prime finders: %v", numFinders)

		finders := make([]<-chan interface{}, numFinders)
		for i := 0; i < numFinders; i++ {
			finders[i] = primeFinder(done, randIntStream)
		}
		for prime := range take(done, fanIn(done, finders...), 10) {
			fmt.Println(prime)
		}

		fmt.Println(time.Since(start))
	*/

	//202 or-done 終了判定のdoneチャネルと入力データのチャネルのどちらかが閉じたら閉じる
	/*
		start := time.Now()
		done := signalForOrDone(10 * time.Second)
		intStream := generatorForOrDone()

		for val := range orDone(done, intStream) {
			fmt.Println(val)
		}

		fmt.Println(time.Since(start))
	*/

	//204 tee チャネル
	/*
		done := make(chan interface{})

		out1, out2 := tee(done, generatorTee(done))
		for v1 := range out1 {
			fmt.Printf("out1:%v , out2: %v \n", v1, <-out2)
		}
	*/

	//206 bridge チャネルを送るチャネルを単一のチャネルとして扱う
	/*
		done := make(chan interface{})
		for v := range bridge(done, generateVals()) {
			fmt.Println(v)
		}
	*/

	//207 context1
	/*
		var wg sync.WaitGroup
		done := make(chan interface{})

		defer close(done)

		//cancel pattern
		//go func() {
		//	time.Sleep(1 * time.Millisecond)
		//	close(done)
		//}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			if isDone, err := shortProcess(done); err != nil {
				fmt.Println("shortProcess: %\n", err)
				fmt.Println(isDone)
				return
			}

			fmt.Println("shortProcess is Done")
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()

			if isDone, err := longProcess(done); err != nil {
				fmt.Printf("longProcess: %v \n", err)
				fmt.Println(isDone)

				return
			}

			fmt.Println("longProcess is Done")
		}()
		wg.Wait()
	*/

	//209 context3
	/*
		var wg sync.WaitGroup

		ctx, cancel := context.WithCancel(context.Background())
		//defer cancel()

		go func() {
			time.Sleep(1 * time.Millisecond)
			cancel()
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			if isDone, err := shortProcessContext(ctx); err != nil {
				fmt.Printf("shortProcess: %v \n", err)
				fmt.Println(isDone)

				cancel()

				return
			}

			fmt.Println("shortProcess is Done")
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()

			if isDone, err := longProcessContext(ctx); err != nil {
				fmt.Printf("longProcess: %v \n", err)
				fmt.Println(isDone)

				cancel()

				return
			}

			fmt.Println("longProcess is Done")
		}()
		wg.Wait()
	*/

	//210 context4
	/*
		var wg sync.WaitGroup
		ctx, cancel := context.WithCancel(context.Background())

		go func() {
			time.Sleep(1 * time.Millisecond)
			cancel()
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			if isDone, err := shortProcessDeadline(ctx); err != nil {
				fmt.Printf("shortProcess: %v \n", err)
				fmt.Println(isDone)

				cancel()

				return
			}

			fmt.Println("shortProcess is Done")
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()

			if isDone, err := longProcessDeadline(ctx); err != nil {
				fmt.Printf("longProcess: %v \n", err)
				fmt.Println(isDone)

				cancel()

				return
			}

			fmt.Println("longProcess is Done")
		}()
		wg.Wait()
	*/

	//211 context5
	/*
		ctx := Set("12345", "abc123")
		Get(ctx)
	*/

	//212 producer-consumer
	/*
		ch := make(chan int)
		var wg = sync.WaitGroup{}

		for i := 1; i < runtime.NumCPU(); i++ {
			go worker(i, ch, &wg)
		}

		for i := 1; i <= 100; i++ {
			wg.Add(1)
			go func(x int) {
				ch <- x * 2
			}(i)
		}
		wg.Wait()
		close(ch)
	*/

	//213 future-promise
	/*
		futureSource := readGoFile("main.go")
		futureFunc := printFunc(futureSource)
		fmt.Println(strings.Join(<-futureFunc, "\n"))
	*/

	//214 heartbeat時間で鼓動を送る
	/*
		done := make(chan interface{})
		const timeout = 2 * time.Second

		heartbeat, results := DoWorkVer1(done, timeout/2)

		for {
			select {
			case _, ok := <-heartbeat:
				if !ok {
					return
				}
				fmt.Println("receive heartbeat")
			case r, ok := <-results:
				if !ok {
					return
				}
				fmt.Printf("result: %v\n", r)
			case <-time.After(timeout):
				fmt.Println("worker gorouttine id dead")
				return
			}
		}
	*/

	//215 heartbeat 仕事ごとに送る
	/*
		done := make(chan interface{})
		defer close(done)

		heartbeat, results := DoWorkVer2(done)
		for {
			select {
			case _, ok := <-heartbeat:
				if ok {
					fmt.Println("Pulse")
				} else {
					return
				}
			case r, ok := <-results:
				if ok {
					fmt.Printf("result: %v\n", r)
				} else {
					return
				}
			}
		}
	*/

	//216 heartbeatを使って再起動するサンプル
	done := make(chan interface{})
	defer close(done)

	dowork, intStream := DoWorkFn(done, 1, 2, -1, 3, 4, 5)
	dowarkWithSteward := newSteward(1*time.Second, dowork)
	dowarkWithSteward(done, 1*time.Second)

	for i := range intStream {
		fmt.Printf("Receieve: %v\n", i)
	}

}
