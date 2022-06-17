package main

import (
	"bufio"
	"fmt"
	"github.com/lionsoul2014/ip2region/binding/golang/xdb"
	"github.com/mitchellh/go-homedir"
	"log"
	"os"
	"strings"
	"time"
)

func printHelp() {
	fmt.Printf("ip2region searcher 2.0\n")
	fmt.Printf("searcher [command] [command options]\n")
	fmt.Printf("Command: \n")
	fmt.Printf("  search    search input test\n")
	fmt.Printf("  bench     search bench test\n")
}

func testSearch() {
	var err error
	var dbFile = ""
	for i := 2; i < len(os.Args); i++ {
		r := os.Args[i]
		if len(r) < 5 {
			continue
		}

		if strings.Index(r, "--") != 0 {
			continue
		}

		var eIdx = strings.Index(r, "=")
		if eIdx < 0 {
			fmt.Printf("missing = for args pair '%s'\n", r)
			return
		}

		switch r[2:eIdx] {
		case "db":
			dbFile = r[eIdx+1:]
		}
	}

	if dbFile == "" {
		fmt.Printf("dbmaker test [command options]\n")
		fmt.Printf("options:\n")
		fmt.Printf(" --db string    ip2region binary xdb file path\n")
		return
	}

	dbPath, err := homedir.Expand(dbFile)
	if err != nil {
		fmt.Printf("invalid xdb file path `%s`: %s", dbFile, err)
		return
	}

	searcher, err := xdb.New(dbPath)
	if err != nil {
		log.Fatalf("failed to create searcher: %s", err.Error())
	}
	defer func() {
		searcher.Close()
		fmt.Printf("searcher test program exited, thanks for trying\n")
	}()

	fmt.Println("ip2region 2.0 searcher test program, type `quit` to exit")
	reader := bufio.NewReader(os.Stdin)
	for {
		fmt.Print("ip2region>> ")
		str, err := reader.ReadString('\n')
		if err != nil {
			log.Fatalf("failed to read string: %s", err)
		}

		line := strings.TrimSpace(strings.TrimSuffix(str, "\n"))
		if len(line) == 0 {
			continue
		}

		if line == "quit" {
			break
		}

		tStart := time.Now()
		region, err := searcher.SearchByStr(line)
		if err != nil {
			fmt.Printf("\x1b[0;31merr:%s\x1b[0m\n", err.Error())
		} else {
			fmt.Printf("\x1b[0;32m{region:%s, took:%s}\x1b[0m\n", region, time.Since(tStart))
		}
	}
}

func testBench() {
	var err error
	var dbFile, srcFile = "", ""
	for i := 2; i < len(os.Args); i++ {
		r := os.Args[i]
		if len(r) < 5 {
			continue
		}

		if strings.Index(r, "--") != 0 {
			continue
		}

		var eIdx = strings.Index(r, "=")
		if eIdx < 0 {
			fmt.Printf("missing = for args pair '%s'\n", r)
			return
		}

		switch r[2:eIdx] {
		case "db":
			dbFile = r[eIdx+1:]
		case "src":
			srcFile = r[eIdx+1:]
		}
	}

	if dbFile == "" || srcFile == "" {
		fmt.Printf("searcher bench [command options]\n")
		fmt.Printf("options:\n")
		fmt.Printf(" --db string     ip2region binary xdb file path\n")
		fmt.Printf(" --src string    source ip text file path\n")
		return
	}

	dbPath, err := homedir.Expand(dbFile)
	if err != nil {
		fmt.Printf("invalid xdb file path `%s`: %s", dbFile, err)
		return
	}

	searcher, err := xdb.New(dbPath)
	defer func() {
		searcher.Close()
	}()

	handle, err := os.OpenFile(srcFile, os.O_RDONLY, 0600)
	if err != nil {
		fmt.Printf("failed to open source text file: %s\n", err)
		return
	}

	var count, tStart = 0, time.Now()
	var scanner = bufio.NewScanner(handle)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		var l = strings.TrimSpace(strings.TrimSuffix(scanner.Text(), "\n"))
		var ps = strings.SplitN(l, "|", 3)
		if len(ps) != 3 {
			fmt.Printf("invalid ip segment line `%s`\n", l)
			return
		}

		sip, err := xdb.CheckIP(ps[0])
		if err != nil {
			fmt.Printf("check start ip `%s`: %s\n", ps[0], err)
			return
		}

		eip, err := xdb.CheckIP(ps[1])
		if err != nil {
			fmt.Printf("check end ip `%s`: %s\n", ps[1], err)
			return
		}

		if sip > eip {
			fmt.Printf("start ip(%s) should not be greater than end ip(%s)\n", ps[0], ps[1])
			return
		}

		mip := xdb.MidIP(sip, eip)
		for _, ip := range []uint32{sip, xdb.MidIP(sip, mip), mip, xdb.MidIP(mip, eip), eip} {
			region, err := searcher.Search(ip)
			if err != nil {
				fmt.Printf("failed to search ip '%s': %s\n", xdb.Long2IP(ip), err)
				return
			}

			// check the region info
			if region != ps[2] {
				fmt.Printf("failed Search(%s) with (%s != %s)\n", xdb.Long2IP(ip), region, ps[2])
				return
			}

			count++
		}
	}

	cost := time.Since(tStart)
	fmt.Printf("Bench finished, {total: %d, took: %s, cost: %d ns/op}\n", count, cost, cost.Nanoseconds()/int64(count))
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
		return
	}

	// set the log flag
	log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
	switch strings.ToLower(os.Args[1]) {
	case "search":
		testSearch()
	case "bench":
		testBench()
	default:
		printHelp()
	}
}