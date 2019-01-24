package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha1"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/r3labs/sse"
)

var apiKey = ""
var keyFile = flag.String("key", "id_rsa", "Path to RSA private key")
var SocketURL = "https://genesis.bitdb.network/s/1FnauZ9aUH2Bex6JzdcV4eNX7oLSSEbxtN/ewogICJ2IjogMywKICAicSI6IHsKICAgICJmaW5kIjogewogICAgICAib3V0LmgxIjogIjgwOGQiIAogICAgICAKICAgIH0KICB9LAogICJyIjogewogICAgImYiOiAiWy5bXSB8IHttc2c6IC5vdXRbXSB8IHNlbGVjdCguYjAub3A/ID09IDEwNikgfCAuczJ9XSIKICB9Cn0="
var QueryURL = "https://genesis.bitdb.network/q/1FnauZ9aUH2Bex6JzdcV4eNX7oLSSEbxtN/ewogICJ2IjogMywKICAicSI6IHsKICAgICJmaW5kIjogewogICAgICAib3V0LmgxIjogIjgwOGQiIAogICAgICAKICAgIH0KICB9LAogICJyIjogewogICAgImYiOiAiWy5bXSB8IHttc2c6IC5vdXRbXSB8IHNlbGVjdCguYjAub3A/ID09IDEwNikgfCAuczJ9XSIKICB9Cn0="

type QueryRequest struct {
	Version  int                 `json:"v"`
	Encoding map[string]string   `json:"e"`
	Query    string              `json:"q"`
	Data     []map[string]string `json:"data"`
}

type TxObject struct {
	Message string `json:"msg"`
}

type QueryResponse struct {
	Unconfirmed []TxObject `json:"u"`
	Confirmed   []TxObject `json:"c"`
}

func main() {
	flag.Parse()
	privKey := getPrivateKey()

	cl := http.Client{}
	req, err := http.NewRequest("GET", QueryURL, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("key", apiKey)
	resp, err := cl.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	//body, err := ioutil.ReadAll(resp.Body)
	query := QueryResponse{}
	err = json.NewDecoder(resp.Body).Decode(&query)
	if err != nil {
		panic(err)
	}
	fmt.Println("Unconfirmed Txes:")
	for _, u := range query.Unconfirmed {
		move := getMove(u.Message, privKey)
		if move == "" {
			continue
		}
		fmt.Printf("Decrypted String: %v\n", move)
	}
	fmt.Println("----------------")

	fmt.Println("Confirmed Txes:")
	for _, c := range query.Confirmed {
		move := getMove(c.Message, privKey)
		if move == "" {
			continue
		}
		fmt.Printf("Decrypted String: %v\n", move)
	}
	//ev := Event{}
	//err = json.NewDecoder(resp.Body).Decode(&ev)

}

func getMove(m string, privKey *rsa.PrivateKey) string {
	message, err := base64.StdEncoding.DecodeString(m)
	if err != nil {
		return ""
	}
	output := []byte{}
	out, err := rsa.DecryptOAEP(sha1.New(), rand.Reader, privKey, message, output)
	if err != nil {
		fmt.Printf("decrypt: %s", err)
		return ""
	}
	return string(out)
}

func getPrivateKey() *rsa.PrivateKey {
	key, err := ioutil.ReadFile(*keyFile)
	if err != nil {
		panic(err)
	}
	block, _ := pem.Decode(key)
	if block == nil {
		panic(fmt.Errorf("bad key data: %s", "not PEM-encoded"))
	}
	priv, err := x509.ParsePKCS1PrivateKey(block.Bytes)
	if err != nil {
		fmt.Printf("bad private key: %s", err)
		panic(err)
	}
	if got, want := block.Type, "RSA PRIVATE KEY"; got != want {
		fmt.Printf("unknown key type %q, want %q", got, want)
		panic(fmt.Errorf("Got invalid Key type"))
	}
	return priv
}
func launchServer() {
	for true {
		events := make(chan *sse.Event)
		client := sse.NewClient("https://genesis.bitdb.network/socket/1FnauZ9aUH2Bex6JzdcV4eNX7oLSSEbxtN/ewogICJ2IjogMywKICAicSI6IHsKICAgICJmaW5kIjogewogICAgICAib3V0LmgxIjogIjgwOGQiIAogICAgICAKICAgIH0KICB9LAogICJyIjogewogICAgImYiOiAiWy5bXSB8IC5vdXRbXSB8IHNlbGVjdCguYjAub3A/ID09IDEwNikgfCAuczJdIgogIH0KfQ==")
		client.SubscribeChan("messages", events)
		client.SubscribeRaw(func(msg *sse.Event) {
			// Got some data!
			fmt.Printf("%#v\n", string(msg.Data))
		})
	}

}
