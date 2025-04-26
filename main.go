package main

import (
	"coding-challenge-skylink/pkg"
	"fmt"
	"log"
)

func main() {
	g, err := pkg.NewGraph("testdata/input.txt")
	if err != nil {
		log.Fatal(err)
	}

	for _, sink := range g.Sinks {
		pathExists := g.BFS(*g.Source, *sink, nil)
		if !pathExists {
			log.Fatalf("no path from source %s to sink %s", g.Source.Id, sink.Id)
		}
	}
	maxFlow, err := g.MaxFlowToMultipleSinks()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("maxFlow:", maxFlow)
}
