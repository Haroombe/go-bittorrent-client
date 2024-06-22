package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/jackpal/bencode-go"
)

type bencodeInfo struct {
	Pieces      string `bencode:"pieces"`
	PieceLength int    `bencode:"piece length"`
	Length      int    `bencode:"length"`
	Name        string `bencode:"name"`
}

type bencodeTorrent struct {
	Announce string      `bencode:"announce"`
	Info     bencodeInfo `bencode:"info"`
}

// Open parses a torrent file
func Open(r io.Reader) (*bencodeTorrent, error) {
	var bto bencodeTorrent
	err := bencode.Unmarshal(r, &bto)
	if err != nil {
		return nil, err
	}
	return &bto, nil
}

func fetchTorrent(torrentPath string) ([]byte, string, error) {
	fmt.Printf("Fetching torrent file from '%s'\n", torrentPath)
	resp, err := http.Get(torrentPath)
	if err != nil {
		return nil, "", err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, "", err
	}

	if strings.Contains(string(body), "not found") {
		return nil, "", fmt.Errorf("failed to fetch torrent from '%s'", torrentPath)
	}

	if _, err := os.Stat("./torrents/"); os.IsNotExist(err) {
		os.Mkdir("./torrents/", os.ModePerm)
	}

	fileName := path.Base(torrentPath)
	filePath := "./torrents/" + fileName

	err = ioutil.WriteFile(filePath, body, 0644)
	if err != nil {
		return nil, "", err
	}

	return body, filePath, nil
}

func openLocalTorrent(torrentPath string) (io.Reader, error) {
	fmt.Printf("Opening torrent file at '%s'\n", torrentPath)
	file, err := os.Open(torrentPath)
	if err != nil {
		return nil, err
	}
	return file, nil
}

func main() {
	if len(os.Args) != 2 {
		fmt.Println("Usage: 'go run torrent_parser.go <path/to/torrentfile.torrent>' or 'go run torrent_parser.go <link_to_torrent.com>'")
		return
	}
	torrentPath := os.Args[1]

	var reader io.Reader
	var filePath string

	if strings.HasPrefix(torrentPath, "http") {
		body, path, err := fetchTorrent(torrentPath)
		if err != nil {
			log.Fatalln(err)
		}
		filePath = path
		reader = bytes.NewReader(body)
	} else if strings.HasSuffix(torrentPath, ".torrent") {
		file, err := openLocalTorrent(torrentPath)
		if err != nil {
			log.Fatalln(err)
		}
		filePath = torrentPath
		reader = file
	} else {
		fmt.Printf("Invalid Argument '%s'. Pass in path to torrent file or magnet link\n", torrentPath)
		return
	}

	torrent, err := Open(reader)
	if err != nil {
		fmt.Println("Error parsing torrent file:", err)
		return
	}

	fmt.Printf("Announce: %s\n", torrent.Announce)
	fmt.Printf("Name: %s\n", torrent.Info.Name)
	fmt.Printf("Piece Length: %d\n", torrent.Info.PieceLength)
	fmt.Printf("Length: %d\n", torrent.Info.Length)

	var viewPieces string
	fmt.Printf("View pieces of torrent '%s'? (y/n): ", torrent.Info.Name)
	fmt.Scanln(&viewPieces)
	if viewPieces == "y" {
		fmt.Printf("Pieces: %s\n", torrent.Info.Pieces)
	}

	if strings.HasPrefix(torrentPath, "http") {
		fmt.Printf("Torrent file saved locally as '%s'\n", filePath)
	}
}
