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
    bto := bencodeTorrent{}
    err := bencode.Unmarshal(r, &bto)
    if err != nil {
        return nil, err
    }
    return &bto, nil
}

func main() {
    // Check if the correct number of arguments is provided
    if len(os.Args) != 3 {
        fmt.Println("Usage: 'go run torrent_parser.go file <path/to/torrentfile.torrent>' or 'go run torrent_parser.go link <link_to_torrent.com>'")
        return
    }
    torrentPath := os.Args[2]
    argType := os.Args[1]

    var reader io.Reader
    var filePath string

    if argType == "link" {
        fmt.Printf("Fetching torrent file from '%s'\n", torrentPath)
        resp, err := http.Get(torrentPath)
        if err != nil {
            log.Fatalln(err)
        }
        defer resp.Body.Close() // Ensure the body is closed when done

        body, err := ioutil.ReadAll(resp.Body)
        if err != nil {
            log.Fatalln(err)
        }

        // Create the ./torrents/ directory if it does not exist
        if _, err := os.Stat("./torrents/"); os.IsNotExist(err) {
            os.Mkdir("./torrents/", os.ModePerm)
        }

        // Extract the file name from the URL
        fileName := path.Base(torrentPath)
        filePath = "./torrents/" + fileName

        // Save the file locally
        err = ioutil.WriteFile(filePath, body, 0644)
        if err != nil {
            log.Fatalln(err)
        }

        reader = bytes.NewReader(body)
    } else if argType == "file" {
        fmt.Printf("Opening torrent file at '%s'\n", torrentPath)
        filePath = torrentPath

        file, err := os.Open(torrentPath)
        if err != nil {
            fmt.Println("Error opening file:", err)
            return
        }
        defer file.Close()
        reader = file
    } else {
        fmt.Println("Invalid argument type. Use 'file' or 'link'.")
        return
    }

    // Parse the torrent file
    torrent, err := Open(reader)
    if err != nil {
        fmt.Println("Error parsing torrent file:", err)
        return
    }

    // Print out the parsed information
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

    // Notify the user where the file is saved
    if argType == "link" {
        fmt.Printf("Torrent file saved locally as '%s'\n", filePath)
    }
}
