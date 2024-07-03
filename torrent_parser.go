package main

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"github.com/jackpal/bencode-go"
	"io"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"io/ioutil"
	"log"
	"net/http"
)

var PEER_ID = []byte("-PC0123-abcdefghijklmno")
var PORT uint16 = 6881

type TrackerResponse struct {
	Peers string `bencode:"peers"`
	// Add other fields as per tracker response structure
}
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

type TorrentFile struct {
	Announce    string
	InfoHash    [20]byte
	PieceHashes [][20]byte
	PieceLength int
	Length      int
	Name        string
}

func (t *TorrentFile) buildTrackerURL() (string, error) {
	base, err := url.Parse(t.Announce)
	if err != nil {
		return "", err
	}
	params := url.Values{
		"info_hash":  []string{fmt.Sprintf("%x", t.InfoHash)},
		"peer_id":    []string{string(PEER_ID[:])},
		"port":       []string{strconv.Itoa(int(PORT))},
		"uploaded":   []string{"0"},
		"downloaded": []string{"0"},
		"compact":    []string{"1"},
		"left":       []string{strconv.Itoa(t.Length)},
	}
	base.RawQuery = params.Encode()
	return base.String(), nil
}

func calculateInfoHash(infoDict bencodeInfo) ([20]byte, error) {
	// fmt.Printf("Calculating info hash of 'info' dict of torrent\n")
	var buf bytes.Buffer
	err := bencode.Marshal(&buf, infoDict)
	if err != nil {
		return [20]byte{}, fmt.Errorf("failed to marshal info: %v", err)
	}

	// Calculate SHA1 hash
	hash := sha1.Sum(buf.Bytes())

	// Convert hash to hexadecimal string
	return hash, nil
}

func toTorrentFile(torrent bencodeTorrent) (*TorrentFile, error) {
	infoHash, err := calculateInfoHash(torrent.Info)
	if err != nil {
		fmt.Println("Error calculating info hash:", err)
		return &TorrentFile{}, err
	}
	pieceHashLength := 20                      // SHA1 hash length
	piecesarray := []byte(torrent.Info.Pieces) // actual pieces in byte array form
	var pieceHashes [][20]byte                 // empty array to be filled with 20 byte sha1 hashes of the pieces
	for i := 0; i < len(piecesarray); i += pieceHashLength {
		var pieceHash [20]byte
		copy(pieceHash[:], piecesarray[i:i+pieceHashLength])
		pieceHashes = append(pieceHashes, pieceHash)
	}
	torrentfile := TorrentFile{
		Announce:    torrent.Announce,
		InfoHash:    infoHash,
		PieceHashes: pieceHashes,
		PieceLength: torrent.Info.PieceLength,
		Length:      torrent.Info.Length,
		Name:        torrent.Info.Name,
	}

	return &torrentfile, nil

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
		fmt.Printf("Invalid Argument '%s'. Pass in path to torrent file or link\n", torrentPath)
		return
	}

	torrent, err := Open(reader)
	if err != nil {
		fmt.Println("Error parsing torrent file:", err)
		return
	}

	if strings.HasPrefix(torrentPath, "http") {
		fmt.Printf("Torrent file saved locally as '%s'\n", filePath)
	}

	fmt.Print("Preparing for tracker announcement\n")
	torrentfile, err := toTorrentFile(*torrent)
	// infoHash, err := calculateInfoHash(torrent.Info)
	if err != nil {
		fmt.Println("Error preparing torrent for announcement:", err)
		return
	}
	fmt.Printf("Announce: %s\n", torrent.Announce)
	fmt.Printf("Name: %s\n", torrent.Info.Name)
	fmt.Printf("Piece Length: %d\n", torrent.Info.PieceLength)
	fmt.Printf("Length: %d\n", torrent.Info.Length)
	fmt.Println("Info Hash:", fmt.Sprintf("%x", torrentfile.InfoHash))
	fmt.Println("Number of pieces after separation:", len(torrentfile.PieceHashes))
	trackerurl, err := torrentfile.buildTrackerURL()
	if err != nil {
		fmt.Println("Error fetching tracker url:", err)
		return
	}
	fmt.Println("Tracker URL:", trackerurl)
	fmt.Println()
	resp, err := http.Get(trackerurl)
	if err != nil {
		fmt.Print("ERROR:")
		log.Fatalln(err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading tracker response:", err)
		return
	}
	fmt.Println("raw tracker response ", resp)

	var trackerResp TrackerResponse
	tracker_urlreader := bytes.NewReader(body)
	fmt.Println("raw tracker response body", body)
	err = bencode.Unmarshal(tracker_urlreader, &trackerResp)
	if err != nil {
		fmt.Println("Error decoding tracker response:", err)
		return
	}

	// Now you can process the peer information
	

}
