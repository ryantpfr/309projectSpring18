//
//  ScoreBoard.swift
//  soccer game
//
//  Created by Mark Schwartz on 3/26/18.
//  Copyright © 2018 MG 6. All rights reserved.
//

import Foundation
import SpriteKit

import Foundation
import SpriteKit

class ScoreBoard{
    
    var redTeamScore: Int
    var blueTeamScore: Int
    var redTeamLabel : SKLabelNode?
    var blueTeamLabel : SKLabelNode?
    
    init(redTeamLabel : SKLabelNode, blueTeamLabel: SKLabelNode)
    {
        redTeamScore = 0
        blueTeamScore = 0
        self.redTeamLabel = redTeamLabel
        self.blueTeamLabel = blueTeamLabel
    }
    func redTeamScored()
    {
        redTeamScore += 1
        redTeamLabel?.text = String(redTeamScore)
    }
    func blueTeamScored()
    {
        blueTeamScore += 1
        blueTeamLabel?.text = String(blueTeamScore)
    }
    
    func forceScore(team1: Int, team2: Int){
        redTeamScore = team1
        blueTeamScore = team2
        redTeamLabel?.text = String(redTeamScore)
        blueTeamLabel?.text = String(blueTeamScore)
    }
    
}




