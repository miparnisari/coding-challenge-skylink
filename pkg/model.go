package pkg

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"syscall"
)

var (
	EdgeRegex      = regexp.MustCompile(`TRANSMISSION: NODE (?P<s>.*) RELAYS (?P<e>.*) UNDER QUOTA (?P<q>.*)`)
	StartNodeRegex = regexp.MustCompile(`ALERT: PRIMARY NODE IS (?P<s>.*)`)
	EndNodesRegex  = regexp.MustCompile(`FINAL ARRIVAL POINTS ARE (?P<e>.*)`)
)

func NewGraph(filename string) (*Network, error) {
	cityGraph := &Network{
		Vertices:  make(map[string]Vertex),
		EdgesFrom: make(map[string][]*Edge),
	}

	f, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("could not open graph file: %w", err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {
			log.Printf("could not close graph file: %v", err)
		}
	}(f)

	fi, err := f.Stat()
	if err != nil {
		return nil, fmt.Errorf("could not stat graph file: %w", err)
	}

	size := fi.Size()
	if size <= 0 || size != int64(int(size)) {
		return nil, fmt.Errorf("invalid graph file size: %d", size)
	}

	data, err := syscall.Mmap(int(f.Fd()), 0, int(size), syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		return nil, fmt.Errorf("could not call mmap: %w", err)
	}

	defer func() {
		if err := syscall.Munmap(data); err != nil {
			log.Printf("could not call munmap: %v", err)
		}
	}()

	scanner := bufio.NewScanner(bytes.NewReader(data))

	for scanner.Scan() {
		line := scanner.Text()

		err := parseLine(line, cityGraph)
		if err != nil {
			return nil, err
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("could not scan read: %w", err)
	}

	return cityGraph, nil
}

func parseLine(line string, cityGraph *Network) error {
	edgeMatch := EdgeRegex.FindStringSubmatch(line)
	if edgeMatch != nil {
		start := edgeMatch[EdgeRegex.SubexpIndex("s")]
		end := edgeMatch[EdgeRegex.SubexpIndex("e")]
		weightStr := edgeMatch[EdgeRegex.SubexpIndex("q")]
		weightInt, err := strconv.Atoi(weightStr)
		if err != nil {
			return fmt.Errorf("could not convert weight to int: %w", err)
		}

		_ = cityGraph.GetOrAddVertex(start)
		_ = cityGraph.GetOrAddVertex(end)
		err = cityGraph.AddEdge(start, end, weightInt)
		if err != nil {
			return err
		}
	} else {
		startMatch := StartNodeRegex.FindStringSubmatch(line)
		if startMatch != nil {
			start := startMatch[StartNodeRegex.SubexpIndex("s")]
			startV := cityGraph.GetOrAddVertex(start)
			cityGraph.Source = &startV
		} else {
			endsMatch := EndNodesRegex.FindStringSubmatch(line)
			if endsMatch != nil {
				ends := endsMatch[EndNodesRegex.SubexpIndex("e")]
				endsSplit := strings.Split(ends, ",")
				for _, end := range endsSplit {
					cleanEnd := strings.TrimSpace(end)
					endV := cityGraph.GetOrAddVertex(cleanEnd)
					cityGraph.Sinks = append(cityGraph.Sinks, &endV)
				}
			} else {
				return fmt.Errorf("could not parse line: %s", line)
			}
		}
	}
	return nil
}
