package cmd

import (
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/ollama/ollama/discover"
	"github.com/ollama/ollama/format"
	"github.com/ollama/ollama/rpc"
	"github.com/spf13/cobra"
)

// RunRPCServer starts an RPC server for distributed inferencing
func RunRPCServer(cmd *cobra.Command, args []string) error {
	// Default parameters
	host := "127.0.0.1"
	port := 50052
	var backendMem int64 = 0 // 0 means use all available memory

	// Parse command line flags
	hostFlag, err := cmd.Flags().GetString("host")
	if err == nil && hostFlag != "" {
		host = hostFlag
	}

	portFlag, err := cmd.Flags().GetInt("port")
	if err == nil && portFlag > 0 && portFlag < 65536 {
		port = portFlag
	}

	memFlag, err := cmd.Flags().GetString("mem")
	if err == nil && memFlag != "" {
		// Parse memory size (in MB)
		memMB, err := strconv.ParseInt(memFlag, 10, 64)
		if err != nil {
			return fmt.Errorf("invalid memory size: %v", err)
		}
		backendMem = memMB * 1024 * 1024 // Convert MB to bytes
	}

	// Print warning if host is not localhost
	if host != "127.0.0.1" && host != "localhost" {
		fmt.Fprintf(os.Stderr, "\n")
		fmt.Fprintf(os.Stderr, "!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!\n")
		fmt.Fprintf(os.Stderr, "WARNING: Host ('%s') is not localhost\n", host)
		fmt.Fprintf(os.Stderr, "         Never expose the RPC server to an open network!\n")
		fmt.Fprintf(os.Stderr, "         This is an experimental feature and is not secure!\n")
		fmt.Fprintf(os.Stderr, "!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!\n")
		fmt.Fprintf(os.Stderr, "\n")
	}

	// Log available GPUs
	gpus := discover.GetGPUInfo()
	for _, gpu := range gpus {
		slog.Debug("detected GPU", "id", gpu.ID, "library", gpu.Library)
	}

	// Create backend
	backend, err := rpc.CreateBackend()
	if err != nil {
		return fmt.Errorf("failed to create backend: %v", err)
	}
	defer backend.Free()

	// Get backend memory
	freeMem, totalMem := rpc.GetBackendMemory(backendMem)

	// Start RPC server
	endpoint := fmt.Sprintf("%s:%d", host, port)
	fmt.Printf("Starting RPC server on %s, backend memory: %s\n", endpoint, format.HumanBytes2(uint64(freeMem)))

	// Set up signal handling for graceful shutdown
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-signals
		fmt.Println("\nShutting down RPC server...")
		// The defer backend.Free() will be called when the function returns
		os.Exit(0)
	}()

	// Start the RPC server (blocking call)
	if err := rpc.StartRPCServer(backend, endpoint, freeMem, totalMem); err != nil {
		return fmt.Errorf("failed to start RPC server: %v", err)
	}

	return nil
}
