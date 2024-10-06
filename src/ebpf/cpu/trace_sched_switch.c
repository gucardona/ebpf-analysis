#include <linux/bpf.h>
#include <linux/sched.h>
#include <linux/bpf_helpers.h>

struct data_t {
    u32 pid;   // Process ID
    u32 cpu;   // CPU ID
    u64 time;  // Timestamp of the switch
};

struct bpf_map_def SEC("maps/events") events = {
    .type = BPF_MAP_TYPE_RINGBUF,
};

SEC("tracepoint/sched/sched_switch")
int trace_sched_switch(struct data_t *data) {
    data->pid = bpf_get_current_pid_tgid() >> 32; // Get PID
    data->cpu = bpf_get_smp_processor_id();       // Get CPU ID
    data->time = bpf_ktime_get_ns();              // Get current timestamp

    // Output the data to the ring buffer
    bpf_perf_event_output(ctx, &events, BPF_F_CURRENT_CPU, data, sizeof(*data));
    return 0;
}

char _license[] SEC("license") = "GPL";
