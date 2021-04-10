package pkg

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"

	"github.com/peterbourgon/diskv"
)

type VolumeMeta struct {
	NodeId string
	Name   string
}

var db = diskv.New(diskv.Options{
	BasePath:     "data",
	Transform:    func(s string) []string { return []string{} },
	CacheSizeMax: 1024 * 1024, // 1MB
})

var lock sync.Mutex

func NewMeta(volumeId, nodeId string) (*VolumeMeta, error) {
	lock.Lock()
	defer lock.Unlock()
	for i := 'b'; i < 'z'; i++ {
		have := false
		for _, meta := range ListMetas() {
			if meta.Name == fmt.Sprintf("vd%c", i) && meta.NodeId == nodeId {
				have = true
			}
		}
		if !have {
			meta := VolumeMeta{
				Name:   fmt.Sprintf("vd%c", i),
				NodeId: nodeId,
			}
			b, _ := json.Marshal(meta)
			db.Write(volumeId, b)
			return &meta, nil
		}
	}
	return nil, errors.New("empty")
}

func GetMeta(volumeId string) (*VolumeMeta, error) {
	b, err := db.Read(volumeId)
	if err != nil {
		return nil, err
	}
	var meta VolumeMeta
	json.Unmarshal(b, &meta)
	return &meta, nil
}

func RemoveMeta(volumeId string) error {
	return db.Erase(volumeId)
}

func ListMetas() []*VolumeMeta {
	var metas []*VolumeMeta
	for key := range db.Keys(nil) {
		b, _ := db.Read(key)
		var meta VolumeMeta
		json.Unmarshal(b, &meta)
		metas = append(metas, &meta)
	}
	return metas
}
