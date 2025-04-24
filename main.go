package main

import (
	"coding-challenge-skylink/pkg"
	"fmt"
	"github.com/dominikbraun/graph/draw"
	"log"
	"os"
)

func main() {
	g, err := pkg.NewGraph("testdata/input.txt")
	if err != nil {
		log.Fatal(err)
	}
	file, _ := os.Create("./mygraph.gv")
	_ = draw.DOT(g.Graph, file)
	maxFlow, err := g.MaxFlowToMultipleSinks()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("maxFlow:", maxFlow)
}
