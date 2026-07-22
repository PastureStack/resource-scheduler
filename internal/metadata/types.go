package metadata

// Host contains the dated metadata fields consumed by the scheduler.
// Unknown JSON fields are intentionally ignored for forward compatibility.
type Host struct {
	Name            string            `json:"name"`
	AgentIP         string            `json:"agent_ip"`
	HostID          int               `json:"host_id"`
	Labels          map[string]string `json:"labels"`
	UUID            string            `json:"uuid"`
	Hostname        string            `json:"hostname"`
	Memory          int64             `json:"memory"`
	MilliCPU        int64             `json:"milli_cpu"`
	LocalStorageMb  int64             `json:"local_storage_mb"`
	EnvironmentUUID string            `json:"environment_uuid"`
}

// Container contains the dated metadata fields used to reconstruct resource
// and host-port reservations.
type Container struct {
	Name                string   `json:"name"`
	Ports               []string `json:"ports"`
	HostUUID            string   `json:"host_uuid"`
	UUID                string   `json:"uuid"`
	State               string   `json:"state"`
	MemoryReservation   int64    `json:"memory_reservation"`
	MilliCPUReservation int64    `json:"milli_cpu_reservation"`
}
