package cpu

import (
	"bytes"
	"encoding/binary"
	"errors"
	"fmt"
	"github.com/cilium/ebpf"
	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/perf"
	"os"
	"syscall"
)

type dataT struct {
	PID  uint32
	CPU  uint32
	Time uint64
}

var reader *perf.Reader

// InitCPUMetricsCollection initialize eBPF program and attach to tracepoint (do this once)
func InitCPUMetricsCollection() error {
	// Load the eBPF program
	spec, err := ebpf.LoadCollectionSpec("trace_sched_switch.o")
	if err != nil {
		return fmt.Errorf("failed to load eBPF program: %v", err)
	}

	// Load into the kernel
	coll, err := ebpf.NewCollection(spec)
	if err != nil {
		return fmt.Errorf("failed to create eBPF collection: %v", err)
	}

	// Attach the eBPF program to the "sched_switch" tracepoint
	if _, err = link.Tracepoint("sched", "sched_switch", coll.Programs["trace_sched_switch"], nil); err != nil {
		return fmt.Errorf("failed to attach tracepoint: %v", err)
	}

	// Create a perf event reader for the ring buffer
	if _, err = perf.NewReader(coll.Maps["events"], os.Getpagesize()); err != nil {
		return fmt.Errorf("failed to create perf reader: %v", err)
	}

	return nil
}

// CollectCPUMetrics continuously collect CPU metrics from the perf ring buffer
func CollectCPUMetrics() (string, error) {
	record, err := reader.Read()
	if err != nil {
		if errors.Is(err, syscall.EINTR) {
			return "", nil // handle interrupt errors gracefully
		}
		return "", fmt.Errorf("failed to read from perf buffer: %v", err)
	}

	// Deserialize the data
	var event dataT
	if err := binary.Read(bytes.NewBuffer(record.RawSample), binary.LittleEndian, &event); err != nil {
		return "", fmt.Errorf("failed to parse event: %v", err)
	}

	// Return the collected metric
	return fmt.Sprintf("PID: %d, CPU: %d, Time: %d ns\n", event.PID, event.CPU, event.Time), nil
}
