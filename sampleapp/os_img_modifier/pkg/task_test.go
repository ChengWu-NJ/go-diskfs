package pkg

import "testing"

func TestSaveToYaml(t *testing.T) {
	task := &ModiTask{
		Image:        `/tmp/dietpi.img`,
		PartitionNum: 1,
		ModiFiles:    make([]*ModiFile, 0),
	}

	fi := &ModiFile{
		FileName: `/boot/dietpi.txt`,
		ReplaceItems: []*ReplaceItem{
			{
				MatchRegexp: `(?m)^\s*AUTO_SETUP_NET_WIFI_ENABLED=.*`,
				ReplaceStr: `AUTO_SETUP_NET_WIFI_ENABLED=1`,
			},
		},
	}
	task.ModiFiles = append(task.ModiFiles, fi)

	fi = &ModiFile{
		FileName: `/boot/dietpi-wifi.txt`,
		ReplaceItems: []*ReplaceItem{
			{
				MatchRegexp: `(?m)^\s*aWIFI_SSID\[0\]=.*`,
				ReplaceStr: `aWIFI_SSID[0]='CABELL111ABC'`,
			},
			{
				MatchRegexp: `(?m)^\s*aWIFI_KEY\[0\]=.*`,
				ReplaceStr: `aWIFI_KEY[0]='KEY111ABC'`,
			},
			{
				MatchRegexp: `(?m)^\s*aWIFI_KEYMGR\[0\]=.*`,
				ReplaceStr: `aWIFI_KEYMGR[0]='WPA-PSK'`,
			},
		},
	}
	task.ModiFiles = append(task.ModiFiles, fi)

	err := task.SaveToYaml(`/tmp/moditask.yaml`)
	if err != nil {
		t.Error(err)
	}
}

func TestCreateModiTaskFromYaml(t *testing.T) {
	task, err := CreateModiTaskFromYaml(`/tmp/moditask.yaml`)
	if err != nil {
		t.Error(err)
	}

	t.Log(task)
}