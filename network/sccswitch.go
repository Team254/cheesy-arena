package network

import (
	"fmt"
	"log"
	"net"
	"time"

	"golang.org/x/crypto/ssh"
)

func runSSHCommands(host, username, password string, commands []string) error {
	config := &ssh.ClientConfig{
		User: username,
		Auth: []ssh.AuthMethod{
			ssh.Password(password),
		},
		HostKeyCallback: ssh.HostKeyCallback(func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			// Dont't verify the host key, just accept it.
			return nil
		}),
		Timeout: 5 * time.Second, // Adjust timeout as needed
	}

	client, err := ssh.Dial("tcp", net.JoinHostPort(host, "22"), config)
	if err != nil {
		return fmt.Errorf("failed to dial SSH: %w", err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return fmt.Errorf("failed to create SSH session: %w", err)
	}
	defer session.Close()

	modes := ssh.TerminalModes{
		ssh.TTY_OP_ISPEED: 14400, // input speed = 14.4kbaud
		ssh.TTY_OP_OSPEED: 14400, // output speed = 14.4kbaud
	}

	if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
		return fmt.Errorf("failed to request pseudo terminal: %w", err)
	}

	stdin, err := session.StdinPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := session.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	stderr, err := session.StderrPipe()
	if err != nil {
		return fmt.Errorf("failed to create stderr pipe: %w", err)
	}

	if err := session.Shell(); err != nil {
		return fmt.Errorf("failed to start shell: %w", err)
	}

	// Execute commands
	for _, cmd := range commands {
		_, err = stdin.Write([]byte(cmd + "\n"))
		if err != nil {
			return fmt.Errorf("failed to write command '%s' to stdin: %w", cmd, err)
		}
		// Add a small delay to allow the command to execute and output to be processed
		time.Sleep(500 * time.Millisecond) // Adjust as needed
	}

	// Send the exit commands
	_, err = stdin.Write([]byte("exit\n"))
	if err != nil {
		return fmt.Errorf("failed to write 'exit' to stdin: %w", err)
	}
	_, err = stdin.Write([]byte("exit\n"))
	if err != nil {
		return fmt.Errorf("failed to write 'exit' to stdin: %w", err)
	}

	// Read output (optional)
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := stdout.Read(buf)
			if err != nil {
				return
			}
			fmt.Print(string(buf[:n]))
		}
	}()

	// Read errors (optional)
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := stderr.Read(buf)
			if err != nil {
				return
			}
			fmt.Fprintf(log.Default().Writer(), "stderr: %s", string(buf[:n]))
		}
	}()

	// Wait for the session to close
	return session.Wait()
}

func Downredscc() {
	host := "10.0.100.48"  // Client IP address
	username := "admin"    // SSH Username
	password := "1234Five" // SSH Password

	commands := []string{
		"config terminal",
		"interface range gigabitEthernet 1/2-4",
		"shutdown",
		"exit",
		"exit",
	}

	err := runSSHCommands(host, username, password, commands)
	if err != nil {
		log.Fatalf("Failed to run commands: %v", err)
	}

	fmt.Println("Successfully executed commands on", host)
}

func Downbluescc() {
	host := "10.0.100.49"  // Client IP address
	username := "admin"    // SSH Username
	password := "1234Five" // SSH Password

	commands := []string{
		"config terminal",
		"interface range gigabitEthernet 1/2-4",
		"shutdown",
		"exit",
		"exit",
	}

	err := runSSHCommands(host, username, password, commands)
	if err != nil {
		log.Fatalf("Failed to run commands: %v", err)
	}

	fmt.Println("Successfully executed commands on", host)
}

func Upredscc() {
	host := "10.0.100.48"  // Client IP address
	username := "admin"    // SSH Username
	password := "1234Five" // SSH Password

	commands := []string{
		"config terminal",
		"interface range gigabitEthernet 1/2-4",
		"no shutdown",
		"exit",
		"exit",
	}

	err := runSSHCommands(host, username, password, commands)
	if err != nil {
		log.Fatalf("Failed to run commands: %v", err)
	}

	fmt.Println("Successfully executed commands on", host)
}

func Upbluescc() {
	host := "10.0.100.49"  // Client IP address
	username := "admin"    // SSH Username
	password := "1234Five" // SSH Password

	commands := []string{
		"config terminal",
		"interface range gigabitEthernet 1/2-4",
		"no shutdown",
		"exit",
		"exit",
	}

	err := runSSHCommands(host, username, password, commands)
	if err != nil {
		log.Fatalf("Failed to run commands: %v", err)
	}

	fmt.Println("Successfully executed commands on", host)
}
