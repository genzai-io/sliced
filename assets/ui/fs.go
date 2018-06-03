package ui_data

import (
	"github.com/elazarl/go-bindata-assetfs"
)

//noinspection ALL
func FS() *assetfs.AssetFS {
	//noinspection GoUnresolvedReference
	return &assetfs.AssetFS{
		Asset:     Asset,
		AssetDir:  AssetDir,
		AssetInfo: AssetInfo,
		Prefix:    "",
	}
}
