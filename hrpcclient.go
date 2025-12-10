package hivego

import (
	"errors"
	"sync"

	"github.com/cfoxon/jsonrpc2client"
)

type NodeStats struct {
	successCount int
	failureCount int
	rollingAvg   float64
}

type HiveRpcNode struct {
	addresses      []string
	currentIndex   int
	nodeStats      []NodeStats
	mutex          sync.RWMutex
	MaxConn        int
	MaxBatch       int
	NoBroadcast    bool
}

type globalProps struct {
	HeadBlockNumber int    `json:"head_block_number"`
	HeadBlockId     string `json:"head_block_id"`
	Time            string `json:"time"`
}

type hrpcQuery struct {
	method string
	params interface{}
}

func NewHiveRpc(addrs []string) *HiveRpcNode {
	return NewHiveRpcWithOpts(addrs, 1, 1)
}

func NewHiveRpcWithOpts(addrs []string, maxConn int, maxBatch int) *HiveRpcNode {
	nodeStats := make([]NodeStats, len(addrs))
	return &HiveRpcNode{
		addresses:    addrs,
		currentIndex: 0,
		nodeStats:    nodeStats,
		MaxConn:      maxConn,
		MaxBatch:     maxBatch,
	}
}

func (h *HiveRpcNode) GetDynamicGlobalProps() ([]byte, error) {
	q := hrpcQuery{method: "condenser_api.get_dynamic_global_properties", params: []string{}}
	res, err := h.rpcExec(q)
	if err != nil {
		return nil, err
	}
	return res, nil
}

func (h *HiveRpcNode) rpcExec(query hrpcQuery) ([]byte, error) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	numNodes := len(h.addresses)
	for i := 0; i < numNodes; i++ {
		index := (h.currentIndex + i) % numNodes
		endpoint := h.addresses[index]

		rpcClient := jsonrpc2client.NewClientWithOpts(endpoint, h.MaxConn, h.MaxBatch)
		jr2query := &jsonrpc2client.RpcRequest{Method: query.method, JsonRpc: "2.0", Id: 1, Params: query.params}
		resp, err := rpcClient.CallRaw(jr2query)
		if err != nil {
			h.nodeStats[index].failureCount++
			h.updateRollingAvg(index)
			continue
		}

		if resp.Error != nil {
			h.nodeStats[index].failureCount++
			h.updateRollingAvg(index)
			continue
		}

		// Check for bad data: if result is empty, consider it bad
		if len(resp.Result) == 0 {
			h.nodeStats[index].failureCount++
			h.updateRollingAvg(index)
			continue
		}

		// Success
		h.nodeStats[index].successCount++
		h.updateRollingAvg(index)
		h.currentIndex = index // Set to last successful node
		return resp.Result, nil
	}

	return nil, errors.New("all API nodes failed")
}

func (h *HiveRpcNode) updateRollingAvg(index int) {
	total := h.nodeStats[index].successCount + h.nodeStats[index].failureCount
	if total > 0 {
		h.nodeStats[index].rollingAvg = float64(h.nodeStats[index].successCount) / float64(total)
	}
}

func (h *HiveRpcNode) rpcExecBatchFast(queries []hrpcQuery) ([][]byte, error) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	numNodes := len(h.addresses)
	for i := 0; i < numNodes; i++ {
		index := (h.currentIndex + i) % numNodes
		endpoint := h.addresses[index]

		rpcClient := jsonrpc2client.NewClientWithOpts(endpoint, h.MaxConn, h.MaxBatch)

		var jr2queries jsonrpc2client.RPCRequests
		for j, query := range queries {
			jr2query := &jsonrpc2client.RpcRequest{Method: query.method, JsonRpc: "2.0", Id: j, Params: query.params}
			jr2queries = append(jr2queries, jr2query)
		}

		resps, err := rpcClient.CallBatchFast(jr2queries)
		if err != nil {
			h.nodeStats[index].failureCount++
			h.updateRollingAvg(index)
			continue
		}

		// Check if any response has error or bad data
		hasError := false
		for _, respBytes := range resps {
			if len(respBytes) == 0 {
				hasError = true
				break
			}
		}
		if hasError {
			h.nodeStats[index].failureCount++
			h.updateRollingAvg(index)
			continue
		}

		// Success
		h.nodeStats[index].successCount++
		h.updateRollingAvg(index)
		h.currentIndex = index

		var batchResult [][]byte
		batchResult = append(batchResult, resps...)
		return batchResult, nil
	}

	return nil, errors.New("all API nodes failed")
}
