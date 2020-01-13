package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"sync"

	wavesplatform "github.com/wavesplatform/go-lib-crypto"
)

var (
	flagExact        = flag.Bool("exact", false, "exact match")
	flagParallelism  = flag.Int("par", runtime.NumCPU(), "parallelism")
	flagPrefixPhrase = flag.String("p", "", "prefix phrase")
	flagSuffixPhrase = flag.String("s", "", "suffix phrase")
)

func main() {
	flag.Parse()

	suffix, prefix := false, false
	if *flagPrefixPhrase != "" {
		prefix = true
		fmt.Printf("search by seed with prefix %s*******\n", *flagPrefixPhrase)
	}
	if *flagSuffixPhrase != "" {
		suffix = true
		fmt.Printf("search by seed with suffix *******%s\n", *flagSuffixPhrase)
	}
	fmt.Printf("searching will splitted %d separate threads\n", *flagParallelism)
	fmt.Printf("searching exact match %t\n", *flagExact)

	done := make(chan struct{}, 1)
	gsyc := &sync.WaitGroup{}
	for i := 0; i < *flagParallelism; i++ {
		if prefix {
			gsyc.Add(1)
			go generateSeedAndCheckPrefix(done, gsyc, *flagExact)
		}

		if suffix {
			gsyc.Add(1)
			go generateSeedAndCheckSuffix(done, gsyc, *flagExact)

		}
	}
	gsyc.Wait()

	fmt.Printf("terminated\n")
}

func generateSeedAndCheckSuffix(done chan struct{}, wg *sync.WaitGroup, exact bool) {
	defer wg.Done()

	if !exact {
		*flagSuffixPhrase = strings.ToLower(*flagSuffixPhrase)
	}

	crypto := wavesplatform.NewWavesCrypto()
	for {
		select {
		case <-done:
			return
		default:
			seed := crypto.RandomSeed()
			addr := crypto.AddressFromSeed(seed, wavesplatform.MainNet)

			var ok bool
			if exact {
				ok = strings.HasSuffix(string(addr), *flagSuffixPhrase)
			} else {
				ok = strings.HasSuffix(strings.ToLower(string(addr)), *flagSuffixPhrase)
			}
			if ok {
				f, err := os.Create(fmt.Sprintf("%s.txt", addr))
				if err != nil {
					log.Fatalf("coundln't create file: %v, but seed here: %s", err, seed)
				}

				_, err = f.WriteString(string(seed))
				if err != nil {
					log.Fatalf("coundln't write file: %v, but seed here: %s", err, seed)
				}

				_ = f.Close()
				done <- struct{}{}
			}
		}
	}
}

func generateSeedAndCheckPrefix(done chan struct{}, wg *sync.WaitGroup, exact bool) {
	defer wg.Done()

	if !exact {
		*flagSuffixPhrase = strings.ToLower(*flagPrefixPhrase)
	}

	crypto := wavesplatform.NewWavesCrypto()
	for {
		select {
		case <-done:
			return
		default:
			seed := crypto.RandomSeed()
			addr := crypto.AddressFromSeed(seed, wavesplatform.MainNet)

			var ok bool
			if exact {
				ok = strings.HasPrefix(string(addr), *flagPrefixPhrase)
			} else {
				ok = strings.HasPrefix(strings.ToLower(string(addr)), *flagPrefixPhrase)
			}
			if ok {
				f, err := os.Create(fmt.Sprintf("%s.txt", *flagSuffixPhrase))
				if err != nil {
					log.Fatalf("coundln't create file: %v, but seed here: %s", err, seed)
				}

				_, err = f.WriteString(string(seed))
				if err != nil {
					log.Fatalf("coundln't write file: %v, but seed here: %s", err, seed)
				}

				_ = f.Close()
				done <- struct{}{}
			}
		}
	}
}
