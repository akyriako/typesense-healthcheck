package typesense_healthcheck

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type HealthCheckClient struct {
	httpClient  *http.Client
	apiKey      string
	apiPort     uint
	peeringPort uint
	protocol    string
	namespace   string
	nodesPath   string
	inCluster   bool
}

func NewHealthCheckClient(cfg Config, inCluster bool) *HealthCheckClient {
	return &HealthCheckClient{
		httpClient:  &http.Client{Timeout: 100 * time.Millisecond},
		apiKey:      cfg.ApiKey,
		protocol:    cfg.Protocol,
		apiPort:     cfg.ApiPort,
		peeringPort: cfg.PeeringPort,
		namespace:   cfg.Namespace,
		nodesPath:   cfg.NodesPath,
		inCluster:   inCluster,
	}
}

func (h *HealthCheckClient) GetNodeStatus(ctx context.Context, node string) (NodeStatus, error) {
	fqdn := h.getNodeFullyQualifiedDomainName(node)
	url := fmt.Sprintf("%s://%s:%d/status", h.protocol, fqdn, h.apiPort)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return NodeStatus{State: ErrorState}, err
	}

	req.Header.Set("x-typesense-api-key", h.apiKey)

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return NodeStatus{State: UnreachableState}, err
	}
	defer resp.Body.Close()

	//if resp.StatusCode != http.StatusOK {
	//	return NodeStatus{State: ErrorState}, fmt.Errorf("error executing request, http status code: %d", resp.StatusCode)
	//}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return NodeStatus{State: ErrorState}, err
	}

	var nodeStatus NodeStatus
	err = json.Unmarshal(body, &nodeStatus)
	if err != nil {
		return NodeStatus{State: ErrorState}, err
	}

	return nodeStatus, nil
}

func (h *HealthCheckClient) GetClusterStatus(nhc map[string]NodesHealthCheck) ClusterStatus {
	leaderNodes := 0
	notReadyNodes := 0
	unreachableNodes := 0
	availableNodes := len(nhc)
	minRequiredNodes := getMinimumRequiredNodes(availableNodes)

	for _, hc := range nhc {
		if hc.NodeStatus.State == LeaderState {
			leaderNodes++
		}

		if hc.NodeStatus.State == NotReadyState {
			notReadyNodes++
		}

		if hc.NodeStatus.State == UnreachableState {
			unreachableNodes++
		}
	}

	if leaderNodes > 1 {
		return ClusterStatusSplitBrain
	}

	if leaderNodes == 0 {
		if availableNodes == 1 {
			return ClusterStatusNotReady
		} // here is setting as not ready even if the single node returns state ERROR

		if unreachableNodes != 0 {
			return ClusterStatusNotReady
		}

		return ClusterStatusElectionDeadlock
	}

	if leaderNodes == 1 {
		if minRequiredNodes > (availableNodes - (notReadyNodes + unreachableNodes)) {
			return ClusterStatusNotReady
		}
		return ClusterStatusOK
	}

	return ClusterStatusNotReady
}

func (h *HealthCheckClient) GetNodeHealth(ctx context.Context, node string) (NodeHealth, error) {
	fqdn := h.getNodeFullyQualifiedDomainName(node)
	url := fmt.Sprintf("%s://%s:%d/health", h.protocol, fqdn, h.apiPort)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return NodeHealth{Ok: false}, err
	}

	resp, err := h.httpClient.Do(req)
	if err != nil {
		return NodeHealth{Ok: false}, nil
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return NodeHealth{Ok: false}, nil
	}

	var nodeHealth NodeHealth
	err = json.Unmarshal(body, &nodeHealth)
	if err != nil {
		return NodeHealth{Ok: false}, nil
	}

	return nodeHealth, nil
}

func (h *HealthCheckClient) GetClusterHealth(ctx context.Context) (ClusterHealthCheck, error) {
	nodes, err := h.getNodes()
	if err != nil {
		return ClusterHealthCheck{
			ClusterStatus: ClusterStatusNotReady,
			ClusterHealth: false,
		}, err
	}

	chk := ClusterHealthCheck{
		ClusterStatus: ClusterStatusNotReady,
		ClusterHealth: false,
		NodesStatus:   map[string]NodesHealthCheck{},
	}
	for _, node := range nodes {
		nodeStatus, _ := h.GetNodeStatus(ctx, node)
		nodeHealth, _ := h.GetNodeHealth(ctx, node)
		chk.NodesStatus[node] = NodesHealthCheck{NodeStatus: nodeStatus, NodeHealth: nodeHealth}
	}

	chk.ClusterStatus = h.GetClusterStatus(chk.NodesStatus)

	healthyNodes := len(chk.NodesStatus)
	minimumRequired := getMinimumRequiredNodes(healthyNodes)
	for _, ns := range chk.NodesStatus {
		if !ns.NodeHealth.Ok {
			healthyNodes--
		}
	}

	chk.ClusterHealth = false
	if healthyNodes >= minimumRequired && chk.ClusterStatus == ClusterStatusOK {
		chk.ClusterHealth = true
	}

	return chk, nil
}

func (h *HealthCheckClient) getNodes() ([]string, error) {
	data, err := os.ReadFile(h.nodesPath)
	if err != nil {
		return nil, fmt.Errorf("reading node list: %w", err)
	}

	raw := strings.TrimSpace(string(data))
	if raw == "" {
		return nil, fmt.Errorf("no nodes found")
	}

	nodes := strings.Split(raw, ",")
	suffix := fmt.Sprintf(":%d:%d", h.peeringPort, h.apiPort)
	for i, p := range nodes {
		nodes[i] = strings.TrimSuffix(p, suffix)
	}

	return nodes, nil
}

func (h *HealthCheckClient) getNodeFullyQualifiedDomainName(node string) string {
	if !h.inCluster {
		return node
	}

	node = strings.Replace(node, fmt.Sprintf(":%d:%d", h.peeringPort, h.apiPort), "", 1)
	fqdn := fmt.Sprintf("%s.%s.svc.cluster.local", node, h.namespace)

	return fqdn
}

func getMinimumRequiredNodes(availableNodes int) int {
	return (availableNodes-1)/2 + 1
}
