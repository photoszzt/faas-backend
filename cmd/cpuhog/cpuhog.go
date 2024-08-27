package main

import (
	"fmt"
	"os"
	"strconv"
)

//https://stackoverflow.com/questions/41079492/how-to-artificially-increase-cpu-usage
func main() {
	threads, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Printf("Cant parse cpuhog string: %v", err)
	}

	done := make(chan int)

	//for i := 0; i < runtime.NumCPU(); i++ {
	for i := 0; i < threads-1; i++ {
		go func() {
			for {
				select {
				case <-done:
					return
				default:
				}
			}
		}()
	}

	for {
		select {
		case <-done:
			return
		default:
		}
	}

	//maybe do a better way to quit? this is unreachable
	//time.Sleep(time.Second * 10)
	//close(done)
}
