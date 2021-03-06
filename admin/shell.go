package admin

import (
	"Stowaway/common"
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"

	"github.com/sirupsen/logrus"
)

func HandleShellToNode(startNodeControlConn net.Conn, nodeID uint32) {
	inputReader := bufio.NewReader(os.Stdin)
	for {
		command, err := inputReader.ReadString('\n')
		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}
		switch command {
		case "exit\n":
			if nodeID == 1 {
				*CliStatus = "startnode"
			} else {
				*CliStatus = "node " + fmt.Sprint(nodeID)
			}
			respCommand, _ := common.ConstructCommand("SHELL", command, nodeID, AESKey)
			startNodeControlConn.Write(respCommand)
			ReadyChange <- true
			IsShellMode <- true
			return
		default:
			respCommand, _ := common.ConstructCommand("SHELL", command, nodeID, AESKey)
			startNodeControlConn.Write(respCommand)
		}
	}
}

func HandleSSHToNode(startNodeControlConn net.Conn, nodeID uint32) {
	inputReader := bufio.NewReader(os.Stdin)
	logrus.Info("Waiting for response,please be patient")
	if conrinueornot := <-SSHSUCCESS; conrinueornot {
		fmt.Print("(ssh mode)>>>")
		for {
			command, err := inputReader.ReadString('\n')
			if err != nil {
				fmt.Println(err)
				os.Exit(1)
			}
			switch command {
			case "exit\n":
				if nodeID == 1 {
					*CliStatus = "startnode"
				} else {
					*CliStatus = "node " + fmt.Sprint(nodeID)
				}
				respCommand, _ := common.ConstructCommand("SSHCOMMAND", command, nodeID, AESKey)
				startNodeControlConn.Write(respCommand)
				ReadyChange <- true
				IsShellMode <- true
				return
			case "\n":
				fmt.Print("(ssh mode)>>>")
			default:
				respCommand, _ := common.ConstructCommand("SSHCOMMAND", command, nodeID, AESKey)
				startNodeControlConn.Write(respCommand)

			}
		}
	} else {
		return
	}
}

func HandleNodeCommand(startNodeControlConn net.Conn, NodeID string) {
	nodeid64, _ := strconv.ParseInt(NodeID, 10, 32)
	nodeID := uint32(nodeid64)
	for {
		AdminCommand := <-ADMINCOMMANDCHAN
		switch AdminCommand[0] {
		case "shell":
			respCommand, err := common.ConstructCommand("SHELL", "", nodeID, AESKey)
			_, err = startNodeControlConn.Write(respCommand)
			if err != nil {
				logrus.Errorf("ERROR OCCURED!: %s", err)
			}
			HandleShellToNode(startNodeControlConn, nodeID)
		case "socks":
			var socksStartData string
			if len(AdminCommand) == 2 {
				socksStartData = fmt.Sprintf("%s:%s:%s", AdminCommand[1], "", "")
			} else if len(AdminCommand) == 3 {
				fmt.Println("Illegal username/password! Try again!")
				ReadyChange <- true
				IsShellMode <- true
				break
			} else {
				socksStartData = fmt.Sprintf("%s:%s:%s", AdminCommand[1], AdminCommand[2], AdminCommand[3])
			}
			respCommand, err := common.ConstructCommand("SOCKS", socksStartData, nodeID, AESKey)
			_, err = startNodeControlConn.Write(respCommand)
			if err != nil {
				logrus.Error("StartNode seems offline")
			}
			if <-NodeSocksStarted {
				go StartSocksServiceForClient(AdminCommand, startNodeControlConn, nodeID)
			}
			ReadyChange <- true
			IsShellMode <- true
		case "stopsocks":
			err := SocksListenerForClient.Close()
			if err != nil {
				logrus.Error("You have never started socks service!")
			}
			respCommand, _ := common.ConstructCommand("SOCKSOFF", " ", nodeID, AESKey)
			_, err = startNodeControlConn.Write(respCommand)
			if err != nil {
				logrus.Error("StartNode seems offline")
			}
			ReadyChange <- true
			IsShellMode <- true
		case "ssh":
			if len(AdminCommand) == 4 {
				go StartSSHService(startNodeControlConn, AdminCommand, nodeID)
				HandleSSHToNode(startNodeControlConn, nodeID)
			} else {
				fmt.Println("Wrong format! Should be ssh [ip:port] [name] [pass]")
				ReadyChange <- true
				IsShellMode <- true
			}
		case "help":
			ShowNodeHelp()
			ReadyChange <- true
			IsShellMode <- true
		case "":
			ReadyChange <- true
			IsShellMode <- true
			continue
		case "exit":
			*CliStatus = "admin"
			ReadyChange <- true
			IsShellMode <- true
			return
		default:
			fmt.Println("Illegal command, enter help to get available commands")
			ReadyChange <- true
			IsShellMode <- true
		}
	}
}
