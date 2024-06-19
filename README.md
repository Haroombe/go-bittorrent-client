## Overview
This ReadME will be a bit different. I'll be documenting my progress, errors, thinking, etc while attempting to build a bittorrent client in GoLang. This is not a guide, more like my brain train through this. Follow at your own will.

## Motivation (06-15-2024)
I randomly stumbled upon the [Build Your Own X](https://github.com/codecrafters-io/build-your-own-x) github repo while doom scrolling. From there I was interested in the [bittorrent client](https://blog.jse.li/posts/torrent/) project, because I may or may not swash my buckles from time to time if yk what I mean. But fr torrenting is a cool idea, and I wanted to learn how it works.

## 06-16-2024
Ofc I just jumped right in, because I'm smart af duh and I already sometimes may or may not torrent stuff. I just read along and copied code. I found out about [Bencode](https://en.wikipedia.org/wiki/Bencode) (`Ben-Code` funny name lol) and how it's the standard structure used by torrent files. It's kind of like JSON but not really. It supports strings, integers, lists, and dictionaries, but there is a weird length prefix before data.

Anyways Bencoded torrent files should be parsed using a bencode parser, but the guide just copied one from somewhere, so same. I also learned about how torrent files are broken into pieces which are hashed into a fix-length binary blobs using the SHA-1 algorithm. Then I committed and go sleep. 

## 06-17-2024
I realized I am a dumbass and I don't understand how torrenting works, even though I learned [Bencode](https://en.wikipedia.org/wiki/Bencode) and the hashed pieces of the file. So I deep dove into how torrenting works. Here is a brief overview that makes sense to me.

Torrenting is just file exchange from peer (dude) to peer (P2P) over the internet. But how do you connect one dude to another? Two main ways, centrally or de-centrally. 

With central discovery, dudes connect to a central server called a tracker that connects you with other dudes who want to upload or download the file. But the connection is P2P, the tracker only helps you find the dudes and connect to the circle jerk.

## 06-18-2024
Decentral discovery is kinda too complex, read about it [here](https://www.pilot.co.za/blog/utorrent-s-dht-explained-understanding-the-power/#:~:text=A%20DHT%20is%20essentially%20a,use%20is%20down%20or%20unavailable.).

Ok, now we are connected and ready to exchange files. How does the file transfer actually work? We have to talk about how the `.torrent` files are actually created first. So to create a torrent file, we basically take our target file and chop it into pieces. Each piece is then hashed using that SHA-1 algorithm from earlier. The cool thing is that the output is fixed length, to be more specific it's 20 bytes. This means that even large files, say a 100 GB file, will have a torrent file significantly smaller. Ok, we back to the topic. After hashing and getting the array of pieces, we hash all the pieces to give us a unique hash called the `info hash`. The info hash is the **unique identifier for the file**. Now we take all this info and put it into the torrent file. When we connect to a tracker, all our info gets broadcasted and we can then join the torrent network for the file we want to download.

But how does the download work? When we are downloading a file via torrent, we download piece by piece. Each piece we get, we check the validity of by hashing with SHA-1 and **checking against the corresponding piece in our torrent file**. If it matches we keep it, if it doesn't we discard ig. Validating is key bc it stops malicious intent and ensures we get what we want. What was cool to me is that during P2P, the pieces we already downloaded are then downloaded by other people, ie we are seeding portions of the file for others at the same time we are downloading. And after we finish downloading, we can continue to seed the file for other people to download. 

PS: I think it is bad etiquette to only download and then just leave, ie **leeching**.

## 06-19-2024