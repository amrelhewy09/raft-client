package main

import (
	"fmt"
	"log"
	"net/rpc"
	"os"
	"strings"
)

type ClientCommandArgs struct {
	Command string
}

type ClientCommandReply struct {
	Result         string
	Err            string
	RedirectLeader string
}
type GetValueArgs struct {
	Key string
}

type GetValueReply struct {
	Value string
	Found bool
}

func sendCommand(target string, command string) {
	fmt.Println("Sending command:", target, command)
	args := ClientCommandArgs{Command: command}
	var reply ClientCommandReply

	client, err := rpc.DialHTTP("tcp", target)
	if err != nil {
		log.Fatalf("❌ failed to connect to %s: %v", target, err)
	}
	defer client.Close()

	err = client.Call("Node.HandleClientCommand", &args, &reply)
	if err != nil {
		log.Fatalf("❌ RPC error: %v", err)
	}

	if reply.Err == "not leader" && reply.RedirectLeader != "" {
		fmt.Printf("🔁 Redirected to leader at %s\n", reply.RedirectLeader)
		sendCommand(reply.RedirectLeader, command)
		return
	}

	fmt.Printf("✅ Command result: %s\n", reply.Result)
}

func getCommand(addr, key string) {
	key = strings.TrimPrefix(key, "GET ")
	args := GetValueArgs{Key: key}
	var reply GetValueReply

	client, err := rpc.DialHTTP("tcp", addr)
	if err != nil {
		fmt.Println("❌ Connection failed:", err)
		return
	}
	defer client.Close()

	err = client.Call("Node.GetValue", &args, &reply)
	if err != nil {
		fmt.Println("❌ RPC failed:", err)
		return
	}

	if reply.Found {
		fmt.Printf("🧠 %s = %s\n", key, reply.Value)
	} else {
		fmt.Printf("🔍 Key %s not found\n", key)
	}
}

func main() {
	if len(os.Args) < 3 {
		fmt.Println("Usage: go run client.go <node address> <command>")
		os.Exit(1)
	}

	addr := os.Args[1]    // e.g., localhost:8080
	command := os.Args[2] // e.g., SET x 5

	if command[:3] == "GET" {
		getCommand(addr, command)
	} else {
		sendCommand(addr, command)
	}
}
