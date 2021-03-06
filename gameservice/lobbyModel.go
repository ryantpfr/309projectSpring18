package main

/*
	LobbyModel.go contains the literal lobby object with functionality for responding to
	lobby-specific packets including sending any responses over a packet out channel.

	Lobby and it's functions are not thread safe and should be run on a single goRoutine or
	have external blockig set up.
 */

import(
	"fmt"
	"time"
)

const DEFAULT_EMOJI = "🐱"

type Lobby struct{
	players [NUMPLAYERS]lobbyPlayer
	size int
	readyToMoveToGameScene bool

	messages []chatMessage
}

type lobbyPlayer struct{
	ready bool
	username string
	emoji string
	connection *playerConnection
}

type chatMessage struct{
	playerNumber int
	message string
}

//non-communication related logic for adding a player to a lobby
func (l *Lobby) addPlayer(newPlayer *waitingPlayer, sendOut func(PacketOut)){
	i := l.size
	l.size += 1

	newPlayer.connection.packetOut <- PacketOut{ data:[]byte{222,byte(l.size-1)},size:2 }

	l.sendPlayerExistingLobbyInfo(newPlayer)

	l.players[i] = lobbyPlayer{
		ready: false,
		username : newPlayer.connection.playerInfo.Nickname,
		emoji: DEFAULT_EMOJI,
		connection: newPlayer.connection,
	}

	l.tellOtherPlayersYouJoined(newPlayer,sendOut)
}

//send everything to date about a lobby
func (l *Lobby) sendPlayerExistingLobbyInfo(newPlayer *waitingPlayer) {
	l.sendAllExistingPlayers(newPlayer.connection.packetOut)
	l.sendAllChatMessagePackets(newPlayer.connection.packetOut)
	l.sendExistingPlayersReady(newPlayer.connection.packetOut)
	l.sendAllCurrentEmojis(newPlayer.connection.packetOut)
}

func (l *Lobby) sendAllChatMessagePackets(to chan<- PacketOut){

	if debug {fmt.Println("messages are",l.messages)}

	for _ , m := range l.messages {
		message := packet203{
			playerNumber: byte(m.playerNumber),
			message: m.message,
		}

		packetOut := PacketOut{
			size: 402,
			data: message.toBytes(),
		}

		to <- packetOut
	}
}

func (l *Lobby) respondTo202(in *PacketIn, sendOut func(PacketOut)) {
	messageIn := ParseBytesTo202(in.data)
	playerNumber := l.playerNumberForConnectionID(in.connectionId)

	if debug {fmt.Println("repeating message",messageIn.message)}


	message := chatMessage{
		playerNumber: playerNumber,
		message: messageIn.message,
	}

	l.messages = append(l.messages, message)

	messageOut := packet203{
		playerNumber: byte(playerNumber),
		message: messageIn.message,

	}

	packetOut := PacketOut{
		size: 402,
		data: messageOut.toBytes(),
		targetIds: l.allConnectionIdsBut(in.connectionId),
	}

	sendOut(packetOut)
}

func (l *Lobby) respondTo200(in *PacketIn, sendOut func(PacketOut) ){
	playerNum := l.playerNumberForConnectionID(in.connectionId)
	if debug {fmt.Println("player",playerNum,"is ready")}
	l.players[playerNum].ready = true

	packet := packet204{numReady:1}
	packetOut := PacketOut{
		size: 2,
		data: packet.toBytes(),
		targetIds: l.allConnectionIdsBut(in.connectionId),
	}
	sendOut(packetOut)

	if l.areAllPlayersReadyForTheGame() {
		l.readyToMoveToGameScene = true
	}
}

func (l *Lobby) respondTo201(in *PacketIn, sendOut func(PacketOut) ){
	playerNum := l.playerNumberForConnectionID(in.connectionId)
	if debug{fmt.Println("player",playerNum,"is unReady")}
	l.players[playerNum].ready = false

	packet := packet205{numUnready:1}
	packetOut := PacketOut{
		size: 2,
		data: packet.toBytes(),
		targetIds: l.allConnectionIdsBut(in.connectionId),
	}
	sendOut(packetOut)

}

func (l *Lobby) respondTo208(in *PacketIn,sendOut func(packetOut PacketOut)){
	playerNum := l.playerNumberForConnectionID(in.connectionId)

	packetIn := ParseBytesTo208(in.data)


	if debug{fmt.Println("player",playerNum,"changed his emoji to",packetIn.Emoji)}

	packet209 := packet209{
		PlayerNumber: byte(playerNum),
		Emoji: packetIn.Emoji,
	}


	packetOut := PacketOut{
		size: 26,
		data: packet209.toBytes(),
		targetIds: l.allConnectionIdsBut(in.connectionId),
	}

	sendOut(packetOut)
}

func (l *Lobby) sendAllExistingPlayers (to chan<- PacketOut){
	for i := 0 ; i < l.size; i++{
		if i == l.size-1{
			continue //don't send players their own username
		}

		packet := packet206{
			playerNumber: i,
			username: l.players[i].username,
		}

		to <- PacketOut{
			size: 82,
			data: packet.toBytes(),
		}
	}
}

func (l *Lobby) sendAllCurrentEmojis(to chan<- PacketOut){
	for i := 0 ; i < l.size; i++{
		if i == l.size-1{
			continue //don't send players their own emoji
		}

		packet := packet209{
			PlayerNumber: byte(i),
			Emoji: l.players[i].emoji,
		}

		to <- PacketOut{
			size: 26,
			data: packet.toBytes(),
		}
	}
}

func (l *Lobby) tellOtherPlayersYouJoined(player *waitingPlayer, sendOut func(PacketOut)){
	packet := packet206{
		playerNumber: l.playerNumberForConnectionID(player.connection.id),
		username: player.connection.playerInfo.Nickname,
	}

	packetOut := PacketOut{
		size: 82,
		data: packet.toBytes(),
		targetIds: l.allConnectionIdsBut(player.connection.id),
	}
	sendOut(packetOut)
}

func (l *Lobby) allConnectionIdsBut(id int) []int{
	rtn := make([]int, l.size-1)
	rtnIndex := 0

	for i := 0; i < l.size ; i+= 1 {
		player := l.players[i]
		if player.connection.id != id{
			rtn[rtnIndex] = player.connection.id
			rtnIndex += 1
		}
	}

	return rtn
}


func (l *Lobby) areAllPlayersReadyForTheGame() bool{
	if l.size != NUMPLAYERS{
		return false
	}

	for _, val := range l.players{
		if !val.ready {
			return false
		}
	}

	return true
}

func (l *Lobby) sendExistingPlayersReady(out chan PacketOut) {
	numReady := 0

	//count players ready
	for i := 0; i < l.size ; i+= 1 {
		if l.players[i].ready {
			numReady += 1
		}
	}

	packet := packet204{numReady: byte(numReady)}
	packetOut := PacketOut{
		size: 2,
		data: packet.toBytes(),
	}

	out <- packetOut
}

func (l *Lobby) respondTo125(in *PacketIn, sendOut func(PacketOut) ){
	if debug {
		fmt.Println("125, AHHHHHHHHHH")
	}

	packetOut := PacketOut{
		size: 2,
		data: []byte{207,byte(in.connectionId)},
		targetIds: l.allConnectionIdsBut(in.connectionId),
	}
	sendOut(packetOut)

	time.Sleep(time.Second)//so that the message actually gets sent, i don't care enough to handle this properly,
	// because there aren't really any bad repercussions unless the server is running extremely slowly

	for i := 0 ; i < l.size; i += 1 {
		l.players[i].connection.disconnect()
	}
}

func (l *Lobby) playerNumberForConnectionID(id int) int{
	for i := 0; i < l.size; i += 1{
		if l.players[i].connection.id == id {
			return i
		}
	}
	return -1
}
