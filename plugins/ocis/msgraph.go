package main

import (
	"encoding/json"

	storageprovider "github.com/cs3org/go-cs3apis/cs3/storageprovider/v0alpha"
	msgraph "github.com/yaegashi/msgraph.go/v1.0"
)

func (p *OcisPlugin) cs3ResourceToDriveItem(res *storageprovider.ResourceInfo) (*msgraph.DriveItem, error) {
	/*
		{
			"value": [
			  {"name": "myfile.jpg", "size": 2048, "file": {} },
			  {"name": "Documents", "folder": { "childCount": 4} },
			  {"name": "Photos", "folder": { "childCount": 203} },
			  {"name": "my sheet(1).xlsx", "size": 197 }
			],
			"@odata.nextLink": "https://..."
		  }
	*/
	size := new(int)
	*size = int(res.Size) // uint64 -> int :boom:

	driveItem := &msgraph.DriveItem{
		BaseItem: msgraph.BaseItem{
			Name: &res.Path,
		},
		Size: size,
	}
	return driveItem, nil
}

func (p *OcisPlugin) formatDriveItems(mds []*storageprovider.ResourceInfo) ([]byte, error) {
	responses := make([]*msgraph.DriveItem, 0, len(mds))
	for i := range mds {
		res, err := p.cs3ResourceToDriveItem(mds[i])
		if err != nil {
			return nil, err
		}
		responses = append(responses, res)
	}

	return json.Marshal(responses)
}
