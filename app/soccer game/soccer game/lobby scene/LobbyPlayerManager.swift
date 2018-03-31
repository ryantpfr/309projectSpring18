//
//  LobbyPlayerManager.swift
//  soccer game
//
//  Created by rtoepfer on 3/31/18.
//  Copyright © 2018 MG 6. All rights reserved.
//

import Foundation
import SpriteKit

class LobbyPlayerManager{

    var players : [LPMPlayer?] = Array(repeating: nil, count: GameScene.maxPlayers)
    var scene : SKScene
    
    init(scene : SKScene){
        self.scene = scene
    }
    
    func addPlayer(playerNumber: Int, username : String ){
        let label = scene.childNode(withName: "Player \(playerNumber) Label") as! SKLabelNode
        label.text = username
        label.isHidden = false
        
        let newPlayer = LPMPlayer(
            playerNumber : playerNumber,
            username : username,
            label : label
        )
        
        players[playerNumber] = newPlayer
    }
    
    func removePlayer(playerNumber : Int){
        self.players[playerNumber]?.label.isHidden = true
        self.players[playerNumber] = nil
        
        //TODO consolidate remaining players
    }
    
    func playerExists(playerNumber: Int) -> Bool{
        return players[playerNumber] != nil
    }
    
    
    struct LPMPlayer{
        var playerNumber : Int
        var username : String
        var label : SKLabelNode
    }
}