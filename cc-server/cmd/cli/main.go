package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/user/cc-server/internal/cli"
)

var (
	serverAddr string
	clientID   string
	command    string
	token      string
)

var rootCmd = &cobra.Command{
	Use:   "cc-cli",
	Short: "C&C Server CLI - Command and Control Interface",
	Long:  `A CLI tool for managing the Command and Control server.`,
}

var registerCmd = &cobra.Command{
	Use:   "register [token]",
	Short: "Register a new client with the server",
	Long:  `Register a new client using a registration token.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		token = args[0]
		registerClient()
	},
}

var listClientsCmd = &cobra.Command{
	Use:   "list-clients",
	Short: "List all registered clients",
	Long:  `List all clients registered with the C&C server.`,
	Run: func(cmd *cobra.Command, args []string) {
		listClients()
	},
}

var sendCmd = &cobra.Command{
	Use:   "send [client_id] [command]",
	Short: "Send a command to a client",
	Long:  `Send a command to a specific client.`,
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		clientID = args[0]
		command = args[1]
		sendCommand()
	},
}

var getCommandsCmd = &cobra.Command{
	Use:   "get-commands [client_id]",
	Short: "Get command history for a client",
	Long:  `Get the command execution history for a specific client.`,
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		clientID = args[0]
		getCommands()
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&serverAddr, "server", "s", "http://localhost:8080", "Server address")
	
	registerCmd.Flags().StringVarP(&serverAddr, "server", "s", "http://localhost:8080", "Server address")
	listClientsCmd.Flags().StringVarP(&serverAddr, "server", "s", "http://localhost:8080", "Server address")
	sendCmd.Flags().StringVarP(&serverAddr, "server", "s", "http://localhost:8080", "Server address")
	getCommandsCmd.Flags().StringVarP(&serverAddr, "server", "s", "http://localhost:8080", "Server address")

	rootCmd.AddCommand(registerCmd)
	rootCmd.AddCommand(listClientsCmd)
	rootCmd.AddCommand(sendCmd)
	rootCmd.AddCommand(getCommandsCmd)
}

func registerClient() {
	fmt.Printf("Registering client with token: %s\n", token)
	// For client registration, we'll need to implement this in a separate client tool
	// The CLI is for server management, not client registration
	fmt.Println("Use the client binary to register a new client, not this CLI tool")
}

func listClients() {
	apiClient := cli.NewAPIClient(serverAddr)
	clients, err := apiClient.ListClients()
	if err != nil {
		fmt.Printf("Error fetching clients: %v\n", err)
		os.Exit(1)
	}

	if len(clients) == 0 {
		fmt.Println("No clients registered")
		return
	}

	fmt.Printf("Found %d client(s):\n", len(clients))
	for _, client := range clients {
		fmt.Printf("- ID: %s, Hostname: %s, IP: %s, Status: %s, Last Seen: %s\n",
			client.ID, client.Hostname, client.IP, client.Status, client.LastSeen)
	}
}

func sendCommand() {
	apiClient := cli.NewAPIClient(serverAddr)
	cmd, err := apiClient.SendCommand(clientID, command)
	if err != nil {
		fmt.Printf("Error sending command: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Command sent successfully:\n")
	fmt.Printf("ID: %s\n", cmd.ID)
	fmt.Printf("Command: %s\n", cmd.Command)
	fmt.Printf("Status: %s\n", cmd.Status)
}

func getCommands() {
	apiClient := cli.NewAPIClient(serverAddr)
	commands, err := apiClient.GetCommands(clientID)
	if err != nil {
		fmt.Printf("Error fetching commands: %v\n", err)
		os.Exit(1)
	}

	if len(commands) == 0 {
		fmt.Printf("No commands found for client %s\n", clientID)
		return
	}

	fmt.Printf("Found %d command(s) for client %s:\n", len(commands), clientID)
	for _, cmd := range commands {
		fmt.Printf("- ID: %s, Command: %s, Status: %s\n",
			cmd.ID, cmd.Command, cmd.Status)
		if cmd.Result != "" {
			fmt.Printf("  Result: %s\n", cmd.Result)
		}
	}
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}