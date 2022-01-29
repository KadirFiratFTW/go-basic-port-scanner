package main

import (
	"fmt"
	"bufio"
	"os"
	"net"
	"strings"
	"strconv"
	"github.com/withmandala/go-log"
	"golang.org/x/sync/semaphore"
	"sync"
	"time"
	"context"
)

var logger = log.New(os.Stderr).WithoutTimestamp()

type PortScanner struct {
	ip string
	start int
	end int
	timeout int
	displayOptions bool
	lock *semaphore.Weighted
}

func main() {
	host := getIPAddress()
	start, end := getPortRange()
	timeout := getTimeout()
	threadCount := getThread()
	displayOptions := getDisplayOption()
	
	portScanner := &PortScanner{
		ip: host,
		start: start,
		end: end,
		timeout: timeout,
		displayOptions: displayOptions,
		lock: semaphore.NewWeighted(int64(threadCount)),
	}

	logger.Info("Starting..");
	time.Sleep(2000)
	portScanner.StartScanner();
}

func ScanPort(ip string, port int, timeout int, displayOptions bool) {
    target := fmt.Sprintf("%s:%d", ip, port)
    conn, err := net.DialTimeout("tcp", target, time.Duration(timeout) * time.Second)
    
    if err != nil {
        if strings.Contains(err.Error(), "too many open files") {
            time.Sleep(time.Duration(timeout) * time.Second)
            ScanPort(ip, port, timeout, displayOptions)
        } else {
			if displayOptions != true {
            	logger.Info(port, "CLOSED");
			}
        }
        return
    }
    
    conn.Close()
    logger.Info(port, "OPEN");
}

func (ps *PortScanner) StartScanner() {
    wg := sync.WaitGroup{}
    defer wg.Wait()
    
    for port := ps.start; port <= ps.end; port++ {       
        wg.Add(1)
        ps.lock.Acquire(context.TODO(), 1)        
		go func(port int) {
            defer ps.lock.Release(1)
            defer wg.Done()
            ScanPort(ps.ip, port, ps.timeout, ps.displayOptions)
        }(port)
    }
}


func getIPAddress() string{
	host := bufio.NewScanner(os.Stdin)
	ip := "127.0.0.1"
	for  {
		fmt.Print("Host (Enter for 127.0.0.1): ");
		host.Scan()
		if host.Text() == "" {
			break;
		}

		if net.ParseIP(host.Text()) != nil {
			ip = host.Text()
			break;
		}
		logger.Warn("Invalid IP address. Please enter valid ip adress..")
	}
	return ip
}

func getPortRange() (int, int){
	portRange := bufio.NewScanner(os.Stdin)
	portStart := 0
	portEnd := 0
	for  {
		fmt.Print("Port Range (0-65535) : ");
		portRange.Scan()
		input := strings.ReplaceAll(portRange.Text(), " ", "")
		splitRange := strings.Split(input, "-")

		if len(splitRange) <= 1 || splitRange[1] == "" {
			logger.Warn("example : 22-443")
			continue
		}

		start, errStart := strconv.Atoi(splitRange[0])
		end, errEnd := strconv.Atoi(splitRange[1])
		
		if errStart != nil || errEnd != nil || start <= 0 || end > 65535{
			logger.Warn("Port range should defined between 0 and 65535. example : 22-443")
			continue
		}
		
		portStart = start
		portEnd = end
		break;
	}
	return portStart, portEnd
}


func getTimeout() int{
	input := bufio.NewScanner(os.Stdin)
	timeout := 1
	for  {
		fmt.Print("Timeout in seconds between requests (Default: 1): ");
		input.Scan()
		if input.Text() == "" {
			break;
		}

		time, err := strconv.Atoi(input.Text())
		if err != nil {
			logger.Warn("Please enter integer value.")
			continue;
		}

		timeout = time
		break;
	}
	return timeout
}

func getThread() int{
	input := bufio.NewScanner(os.Stdin)
	thread := 4
	for  {
		fmt.Print("Threads (Default: 4): ");
		input.Scan()
		if input.Text() == "" {
			break;
		}

		t, err := strconv.Atoi(input.Text())
		if err != nil {
			logger.Warn("Please enter integer value.")
			continue;
		}

		thread = t
		break;
	}
	return thread
}

func getDisplayOption() bool{
	input := bufio.NewScanner(os.Stdin)
	option := true
	for  {
		fmt.Print("Display only open ports (y (default) / n): ");
		input.Scan()
		if input.Text() == "" || input.Text() == "n" || input.Text() == "n" {
			break;
		}

		if len(input.Text()) > 1 || len(input.Text()) < 1 {
			logger.Warn("Please enter 'y' for yes, 'n' for no.", input.Text())
			continue;
		}

		break;
	}
	return option
}
