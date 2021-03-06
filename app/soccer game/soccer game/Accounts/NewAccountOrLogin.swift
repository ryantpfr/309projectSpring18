//
//  NewAccountOrLogin.swift
//  soccer game
//
//  Created by Mark Schwartz on 2/13/18.
//  Copyright © 2018 MG 6. All rights reserved.
//
import SpriteKit
import UIKit
import FacebookCore
import FacebookLogin
import FBSDKLoginKit
import Alamofire


class NewAccountOrLogin: SKScene{
    
    
    var back : SKNode?
    var logout: SKNode?
    var viewController: UIViewController?
    var logOutLabel: SKLabelNode?
    
    var GamesPlayedLabel : SKLabelNode?
    var GamesWonLabel : SKLabelNode?
    var GoalsScoredLabel : SKLabelNode?
    var GoalRankLabel : SKLabelNode?
    var WinRankLabel : SKLabelNode?
    
    var isLoggedIn = AccessToken.current != nil
    var loginButton = LoginButton(readPermissions: [ .publicProfile ])
    
    override func didMove(to view: SKView) {
        print("got to accounts")
        
        //get scene subnodes
        self.back = self.childNode(withName: "Back Button")
        self.GamesPlayedLabel = self.childNode(withName: "GamesPlayedLabel") as? SKLabelNode
        self.GamesWonLabel = self.childNode(withName: "GamesWonLabel") as? SKLabelNode
        self.GoalsScoredLabel = self.childNode(withName: "GoalsScoredLabel") as? SKLabelNode
        self.GoalRankLabel = self.childNode(withName: "GoalRankLabel") as? SKLabelNode
        self.WinRankLabel = self.childNode(withName: "WinRankLabel") as? SKLabelNode
        
        
        let screenSize:CGRect = UIScreen.main.bounds
        let screenHeight = screenSize.height //real screen height
        //let's suppose we want to have 10 points bottom margin
        let newCenterY = screenHeight - loginButton.frame.height - 10
        let newCenter = CGPoint(x: view.center.x,y:  newCenterY)
     
        
        loginButton.center = newCenter//view.center
        view.addSubview(loginButton)
        
        sendCRUDServiceStatsRequest(FBToken: AccessToken.current!.userId!)
        
    }
    
    override func touchesBegan(_ touches: Set<UITouch>, with event: UIEvent?)
    {
        
        
        //if necesarry nodes and one touch exist
        if let t = touches.first ,let back = self.back
        {
            let point = t.location(in: self)
            
            //see if touch contains first
            if (back.contains(point))
            {
                self.moveToScene(.mainMenu)
                loginButton.removeFromSuperview()
            }
          
                
                
    
        }
    }
        
    
    
    func fadeInLabel(label : SKLabelNode?){
        if let nonOptLabel = label{
            nonOptLabel.alpha = 0.0
            nonOptLabel.run(SKAction.fadeIn(withDuration: 2.0))
        }
    }
    
    func sendCRUDServiceStatsRequest(FBToken : String)
    {
        
        let requestURL = "http://\(CommunicationProperties.crudServiceHost):\(CommunicationProperties.crudServicePort)/player/stats"
        
        let headers: HTTPHeaders = [
            "FacebookID": (AccessToken.current?.userId)!,
            ]
        
        print("\n")
        print("\n")
        
        print("Getting stats with CRUD Service at URL",requestURL,"\nwith headers : \(headers)")
        
        //.responseJSON(completionHandler: statsRequestResponse(_:))
        //Alamofire.request(requestURL , method : .get , headers : headers).responseString(completionHandler: statsRequestHandler(_:))
       
        
        Alamofire.request(requestURL , method : .get , headers : headers)
            .responseJSON { response in
               
                //to get status code
                if let status = response.response?.statusCode {
                    switch(status){
                    case 200:
                        print("success")
                    default:
                        print("error with response status: \(status)")
                    }
                }
                //to get JSON return value
                if let result = response.result.value {
                    let JSON = result as! NSDictionary
                    
                    print(JSON)
                    
                    
                    //let blog = try? JSONDecoder().decode(Profile.self, from: JSON as! Data)
                    //self.buildStatsNodes(player: blog!)
                    
                    //let playerStats = JSON["Profile"] as! Profile
                    //self.buildStatsNodes(player: playerStats)
                }
                
        }
        

        
    }
  
    //gives the labels the correct strings/values
    func buildStatsNodes(player: Profile)
    {
        print(player.gamesplayed)
    }
    struct Profile: Decodable {
        /*var name: String
        var points: Int
        var description: String?*/
        
        var gamesplayed:Int
        var gameswon: Int
        var goalsscored: Int
        var id : Int
        var lastavatar:String
        var nickname:String
        var rankscore : Int
        var rankwin :Int
    }
    
}

