package typesense_healthcheck

type NodeState string

const (
	LeaderState      NodeState = "LEADER"
	FollowerState    NodeState = "FOLLOWER"
	CandidateState   NodeState = "CANDIDATE"
	NotReadyState    NodeState = "NOT_READY"
	ErrorState       NodeState = "ERROR"
	UnreachableState NodeState = "UNREACHABLE"
)

type NodeStatus struct {
	CommittedIndex int       `json:"committed_index"`
	QueuedWrites   int       `json:"queued_writes"`
	State          NodeState `json:"state"`
}

type ClusterStatus string

const (
	ClusterStatusOK               ClusterStatus = "OK"
	ClusterStatusSplitBrain       ClusterStatus = "SPLIT_BRAIN"
	ClusterStatusNotReady         ClusterStatus = "NOT_READY"
	ClusterStatusElectionDeadlock ClusterStatus = "ELECTION_DEADLOCK"
)

type NodeHealthResourceError string

const (
	OutOfMemory NodeHealthResourceError = "OUT_OF_MEMORY"
	OutOfDisk   NodeHealthResourceError = "OUT_OF_DISK"
)

type NodeHealth struct {
	Ok            bool                     `json:"ok"`
	ResourceError *NodeHealthResourceError `json:"resource_error,omitempty"`
}

type NodesHealthCheck struct {
	NodeStatus NodeStatus `json:"node_status"`
	NodeHealth NodeHealth `json:"node_health"`
}

type ClusterHealthCheck struct {
	ClusterStatus ClusterStatus               `json:"cluster_status"`
	ClusterHealth bool                        `json:"cluster_health"`
	NodesStatus   map[string]NodesHealthCheck `json:"nodes_health_check"`
}
