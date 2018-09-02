package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/davecgh/go-spew/spew"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

//Block ...Block
type Block struct {
	Index     int
	Timestamp time.Time
	BPM       int
	Hash      string
	PrevHash  string
}

// Message ...
type Message struct {
	BPM int
}

// BlockChain ...BlockChain
var BlockChain []Block

func (b *Block) calculateHash() string {
	bVal := b
	record := string(bVal.Index) + bVal.Timestamp.String() + string(bVal.BPM) + string(bVal.PrevHash)
	h := sha256.New()
	h.Write([]byte(record))
	hashed := h.Sum(nil)
	return hex.EncodeToString(hashed)
}

func generateBlock(oldBlock *Block, BPM int) (Block, error) {
	t := time.Now()
	b := Block{oldBlock.Index + 1, t, BPM, "", oldBlock.Hash}
	b.Hash = b.calculateHash()
	return b, nil
}

func isBlockValid(newBlock, oldBlock Block) bool {
	if oldBlock.Index+1 != newBlock.Index {
		return false
	}
	if oldBlock.Hash != newBlock.PrevHash {
		return false
	}
	if newBlock.Hash != newBlock.calculateHash() {
		return false
	}
	return true
}

func replaceChain(newBlocks []Block) {
	if len(newBlocks) > len(BlockChain) {
		BlockChain = newBlocks
	}
}

func handleGetBlockchain(w http.ResponseWriter, r *http.Request) {
	bytes, err := json.MarshalIndent(BlockChain, "", "  ")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	io.WriteString(w, string(bytes))
}

func handleWriteBlock(w http.ResponseWriter, r *http.Request) {
	var m Message
	decoder := json.NewDecoder(r.Body)
	err := decoder.Decode(&m)
	if err != nil {
		log.Println("Json parsin failed", err)
	}
	defer r.Body.Close()
	newBlock, err := generateBlock(&BlockChain[len(BlockChain)-1], m.BPM)
	if err != nil {
		log.Println("Block creation failed", err)
	}
	spew.Dump(newBlock)
	oldBlock := &BlockChain[len(BlockChain)-1]
	if isBlockValid(newBlock, *oldBlock) {
		newBlockChain := append(BlockChain, newBlock)
		spew.Dump(newBlockChain)
		replaceChain(newBlockChain)
	}
	log.Println("Done creating the block !!!")
}

func makeMuxRouter() http.Handler {
	muxRouter := mux.NewRouter()
	muxRouter.HandleFunc("/", handleGetBlockchain).Methods("GET")
	muxRouter.HandleFunc("/", handleWriteBlock).Methods("POST")
	return muxRouter
}

func run() error {
	mux := makeMuxRouter()
	httpAddr := os.Getenv("ADDR")
	log.Println("Listening on ", os.Getenv("ADDR"))
	s := &http.Server{
		Addr:           ":" + httpAddr,
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	if err := s.ListenAndServe(); err != nil {
		return err
	}

	return nil
}

func main() {
	log.Println("Hello World")
	godotenv.Load()
	t := time.Now()
	genesisBlock := Block{}
	genesisBlock = Block{0, t, 0, "", ""}
	genesisBlock.Hash = genesisBlock.calculateHash()
	spew.Dump(genesisBlock)
	BlockChain = append(BlockChain, genesisBlock)
	log.Fatal(run())
}
