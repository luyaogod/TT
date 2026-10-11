package fep

import (
	"encoding/json"
	"sort"
)

// statusJSON reports the live connections to a control client, so commands in
// other processes can see connection ids and the next `_om` numbers without
// holding any state themselves.
func (p *Proxy) statusJSON() string {
	infos := p.Conns()
	sort.Slice(infos, func(i, j int) bool { return infos[i].ID < infos[j].ID })
	type row struct {
		ID      int    `json:"id"`
		Seq     int    `json:"seq"`
		Alive   bool   `json:"alive"`
		Remote  string `json:"remote,omitempty"`
		BytesRT int64  `json:"bytesRT,omitempty"`
		BytesFE int64  `json:"bytesFE,omitempty"`
	}
	out := struct {
		Conns []row `json:"conns"`
	}{Conns: make([]row, 0, len(infos))}
	for _, ci := range infos {
		out.Conns = append(out.Conns, row{ID: ci.ID, Seq: ci.Seq, Alive: ci.Alive,
			Remote: ci.Remote, BytesRT: ci.BytesRT, BytesFE: ci.BytesFE})
	}
	b, err := json.Marshal(out)
	if err != nil {
		return `{"conns":[]}`
	}
	return string(b)
}
