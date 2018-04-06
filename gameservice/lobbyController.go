package main

import (
	"fmt"
	"sync"
)

type LobbyController struct {
	l Lobby

	mmm *matchMakingModel
	packetIn chan PacketIn
	packetOut chan PacketOut

	disperser lobbyDisperser

	packetRouterMap map[byte]func(*PacketIn, chan<- PacketOut)
}

type lobbyDisperser struct{
	mut sync.Mutex
	connections map[int]chan<- PacketOut
}

func startLobby(mmm *matchMakingModel) {
	fmt.Println("startingn new lobby")

	lc := LobbyController{}
	lc.mmm = mmm
	lc.packetIn = make(chan PacketIn, 50)
	lc.packetOut = make(chan PacketOut, 50)

	lc.l = lc.makeLobby()
	lc.buildPacketMap()

	lc.disperser.connections = make(map[int]chan<- PacketOut)
	go lc.runLobbyDispersion()

	lobbyFull := false
	for !lobbyFull { //listen for players and packets
		select {
		case newPlayer := <-lc.mmm.playerChan:
			lc.addSinglePlayer(newPlayer)
			lobbyFull = lc.l.size == NUMPLAYERS
		case packet := <-lc.packetIn:
			lc.handleSinglePacket(packet)
		}
	}

	fmt.Println("lobby is full")

	for packet := range lc.packetIn { //just listen for players
		lc.handleSinglePacket(packet)

		//TODO make logic to break loop to end lobby and GORoutine if all players have disconnected
	}

	fmt.Println("Lobby closing")
}

func (lc *LobbyController) addSinglePlayer(newPlayer *waitingPlayer) {
	lc.mmm.decrementOpenSpaces()
	fmt.Println("lobby added player with id", newPlayer.connection.id, ", current size is",lc.l.size+1)

	newPlayer.connection.SetNewPacketInChannel(lc.packetIn)

	lc.startDispersingToConnection(newPlayer.connection)

	lc.l.addPlayer(newPlayer,lc.packetOut)
}

func (lc *LobbyController) handleSinglePacket(packet PacketIn){
	fmt.Println("lc got a packet", packet.data)

	packetByte := packet.data[0]
	lc.packetRouterMap[packetByte](&packet,lc.packetOut)
}

func (lc *LobbyController) runLobbyDispersion() {
	for packet := range lc.packetOut {
		lc.disperser.mut.Lock()
		fmt.Println("dispersion map is",lc.disperser.connections)
		fmt.Println("sending to",packet.targetIds)
		for _, id := range packet.targetIds{
			lc.disperser.connections[id] <- packet
		}
		lc.disperser.mut.Unlock()
	}
}

func (lc *LobbyController) startDispersingToConnection(connection *playerConnection){
	lc.disperser.mut.Lock()
	lc.disperser.connections[connection.id] = connection.packetOut
	lc.disperser.mut.Unlock()
}

func (lc *LobbyController) makeLobby() (l Lobby) {
	l.size = 0
	l.messages = []chatMessage{}
	l.players = [NUMPLAYERS]lobbyPlayer{}

	return
}

//builds a map of packet types to handler functions
func (lc *LobbyController) buildPacketMap() {
	packetMap := map[byte]func(*PacketIn, chan<- PacketOut){}

	packetMap[125] = my125Stub
	packetMap[202] = lc.l.respondTo202

	lc.packetRouterMap = packetMap
}

func my125Stub(in *PacketIn, out chan<- PacketOut){
	fmt.Println("125, AHHHHHHHHHH")
}
