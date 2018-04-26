package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"
)

var Chain Blockchain

type Blockchain []*Block

type Block struct {
	Index     int
	Timestamp string
	Msg       Message
	Hash      string
	PrevHash  string
}

type Message struct {
	BPM int
}

func (m *Message) String() string {
	if m == nil {
		return ""
	}
	return strconv.Itoa(m.BPM)
}

// 计算block的hash
func (block *Block) calculateHash() (v string, err error) {
	record := strconv.Itoa(block.Index) + block.Timestamp + block.Msg.String() + block.PrevHash
	hash := sha256.New()
	_, err = hash.Write([]byte(record))
	if err != nil {
		return "", err
	}
	hashed := hash.Sum(nil)
	v = hex.EncodeToString(hashed)
	return v, nil
}

func generateBlock(block *Block, BPM int) (newBlock *Block, err error) {
	newBlock = new(Block)
	newBlock.Index = block.Index + 1
	newBlock.Timestamp = time.Now().Format("2006-01-02 15:04:05")
	newBlock.Msg = Message{BPM: BPM}
	newBlock.PrevHash = block.Hash
	newBlock.Hash, err = newBlock.calculateHash()
	return
}

//校验新区快是否是合法的区块
func isValidBlock(newBlock, preBlock *Block) bool {
	if newBlock.Index != preBlock.Index+1 {
		return false
	}

	if newBlock.PrevHash != preBlock.Hash {
		return false
	}

	v, err := newBlock.calculateHash()
	if err != nil {
		return false
	}

	if v != newBlock.Hash {
		return false
	}
	return true
}

func repaceChain(chain Blockchain) {
	if chain == nil {
		return
	}

	if len(chain) > len(Chain) {
		Chain = chain
	}
}

func init() { //初始化创世区块
	Chain = make(Blockchain, 0)
	genesisBlock := &Block{}
	genesisBlock.Index = 0
	genesisBlock.Timestamp = time.Now().Format("2006-01-02 15:04:05")
	genesisBlock.Msg = Message{}
	Chain = append(Chain, genesisBlock)
}

func responseWithError(w http.ResponseWriter, err error, code int) {
	http.Error(w, err.Error(), code)
}

func handleBlockchain(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		data, err := json.MarshalIndent(Chain, "", "\t")
		if err != nil {
			responseWithError(w, err, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write(data)
		return
	}

	if r.Method == http.MethodPost {
		var msg Message
		decoder := json.NewDecoder(r.Body)
		if err := decoder.Decode(&msg); err != nil {
			responseWithError(w, err, http.StatusBadRequest)
			return
		}
		// 进行挖矿处理  若是成功则生成新的区块追加到区块链上
		newBlock, err := generateBlock(Chain[len(Chain)-1], msg.BPM)
		if err != nil {
			responseWithError(w, err, http.StatusInternalServerError)
			return
		}
		//校验新生成的区块
		if isValidBlock(newBlock, Chain[len(Chain)-1]) {
			Chain = append(Chain, newBlock) //追加到区块链上
		}
		data, err := json.MarshalIndent(newBlock, "", "\t")
		if err != nil {
			responseWithError(w, err, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(data)
		return
	}
}

func prettyJson(b []byte) ([]byte, error) {
	var buf bytes.Buffer
	err := json.Indent(&buf, b, "", "\t")
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func main() {
	http.HandleFunc("/", handleBlockchain)
	server := &http.Server{
		Addr:           ":9000",
		Handler:        nil,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
	err := server.ListenAndServe()
	if err != nil {
		fmt.Errorf("Fatal error:%s", err.Error())
	}
}
