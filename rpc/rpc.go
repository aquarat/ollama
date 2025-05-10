// Package rpc provides functionality for the Ollama RPC server
package rpc

import (
	"fmt"
	"log/slog"
	"net"
	"runtime"
	"time"
)

// BackendHandle represents a handle to a backend (CPU, CUDA, Metal, etc.)
type BackendHandle struct {
	// In a real implementation, this would be a pointer to a C struct
	// For now, we'll just use a string to identify the backend type
	backendType string
}

// CreateBackend creates a backend based on available hardware
// It tries to create a GPU backend first (CUDA, Metal, etc.) and falls back to CPU if none is available
func CreateBackend() (*BackendHandle, error) {
	// In a real implementation, this would call into llama.cpp to create a backend
	// For now, we'll just detect the platform and return a placeholder

	var backendType string

	// Simple platform detection
	switch runtime.GOOS {
	case "darwin":
		// On macOS, we'd use Metal if available
		backendType = "Metal"
	case "windows", "linux":
		// On Windows/Linux, we'd use CUDA if available
		backendType = "CUDA"
	default:
		// Default to CPU
		backendType = "CPU"
	}

	slog.Info("created backend", "type", backendType)
	return &BackendHandle{backendType: backendType}, nil
}

// Free releases the resources associated with the backend
func (b *BackendHandle) Free() {
	// In a real implementation, this would call into llama.cpp to free the backend
	slog.Info("freed backend", "type", b.backendType)
}

// GetBackendMemory returns the free and total memory for the backend
// If requestedMem is greater than 0, it returns that value for both free and total memory
func GetBackendMemory(requestedMem int64) (int64, int64) {
	if requestedMem > 0 {
		return requestedMem, requestedMem
	}

	// In a real implementation, this would call into llama.cpp to get the memory info
	// For now, we'll just return a reasonable default (4GB)
	var freeMem, totalMem int64

	// Get system memory info
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Use 80% of available memory as a reasonable default
	totalMem = int64(m.Sys)
	freeMem = totalMem * 8 / 10

	return freeMem, totalMem
}

// StartRPCServer starts the RPC server with the given backend
// This is a blocking call that will run until the server is stopped
func StartRPCServer(backend *BackendHandle, endpoint string, freeMem, totalMem int64) error {
	if backend == nil {
		return fmt.Errorf("invalid backend")
	}

	// Parse endpoint to validate it
	_, _, err := net.SplitHostPort(endpoint)
	if err != nil {
		return fmt.Errorf("invalid endpoint: %v", err)
	}

	slog.Info("starting RPC server",
		"endpoint", endpoint,
		"backend", backend.backendType,
		"free_memory", freeMem,
		"total_memory", totalMem)

	// In a real implementation, this would call into llama.cpp to start the RPC server
	// For now, we'll just simulate a running server with a simple TCP listener

	// Create a TCP listener
	listener, err := net.Listen("tcp", endpoint)
	if err != nil {
		return fmt.Errorf("failed to start listener: %v", err)
	}
	defer listener.Close()

	fmt.Printf("RPC server started on %s with %s backend\n", endpoint, backend.backendType)
	fmt.Printf("Memory: %d MB free / %d MB total\n", freeMem/(1024*1024), totalMem/(1024*1024))
	fmt.Println("Press Ctrl+C to stop the server")

	// Accept connections
	for {
		conn, err := listener.Accept()
		if err != nil {
			slog.Error("error accepting connection", "error", err)
			continue
		}

		// Handle connection in a goroutine
		go handleConnection(conn)
	}
}

// handleConnection handles a single RPC connection
func handleConnection(conn net.Conn) {
	defer conn.Close()

	// Set a timeout for reading
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))

	// Read command (1 byte)
	buf := make([]byte, 1)
	_, err := conn.Read(buf)
	if err != nil {
		slog.Error("error reading command", "error", err)
		return
	}

	// Process command
	cmd := buf[0]
	switch cmd {
	case 10: // Get memory info
		// Read input size (8 bytes)
		sizeBuf := make([]byte, 8)
		_, err := conn.Read(sizeBuf)
		if err != nil {
			slog.Error("error reading input size", "error", err)
			return
		}

		// In a real implementation, we would process the command and return the result
		// For now, we'll just return some placeholder values

		// Get system memory info
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		freeMem := int64(m.Sys * 8 / 10)
		totalMem := int64(m.Sys)

		// Write reply size (8 bytes)
		replySize := int64(16) // 8 bytes for free memory + 8 bytes for total memory
		replySizeBuf := make([]byte, 8)
		for i := 0; i < 8; i++ {
			replySizeBuf[i] = byte(replySize >> (i * 8))
		}
		_, err = conn.Write(replySizeBuf)
		if err != nil {
			slog.Error("error writing reply size", "error", err)
			return
		}

		// Write free memory (8 bytes)
		freeMemBuf := make([]byte, 8)
		for i := 0; i < 8; i++ {
			freeMemBuf[i] = byte(freeMem >> (i * 8))
		}
		_, err = conn.Write(freeMemBuf)
		if err != nil {
			slog.Error("error writing free memory", "error", err)
			return
		}

		// Write total memory (8 bytes)
		totalMemBuf := make([]byte, 8)
		for i := 0; i < 8; i++ {
			totalMemBuf[i] = byte(totalMem >> (i * 8))
		}
		_, err = conn.Write(totalMemBuf)
		if err != nil {
			slog.Error("error writing total memory", "error", err)
			return
		}

	default:
		slog.Error("unknown command", "cmd", cmd)
	}
}
