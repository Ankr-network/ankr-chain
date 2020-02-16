package common

import "testing"

func TestNewABIUtil(t *testing.T) {
	abi := NewABIUtil("[{\"name\":\"init\",\"inputs\":[],\"outputs\":{\"type\":\"string\"},\"type\":[\"action\"]},{\"name\":\"AddToken\",\"inputs\":[{\"name\":\"symbol\",\"type\":\"string\"},{\"name\":\"contractAddr\",\"type\":\"string\"}],\"outputs\":{\"type\":\"bool\"},\"type\":[\"action\"]},{\"name\":\"Withdraw\",\"inputs\":[{\"name\":\"symbol\",\"type\":\"string\"},{\"name\":\"toAddr\",\"type\":\"string\"},{\"name\":\"amount\",\"type\":\"string\"}],\"outputs\":{\"type\":\"int\"},\"type\":[\"action\",\"event\"]},{\"name\":\"ChangeOwner\",\"inputs\":[{\"name\":\"ownerAddr\",\"type\":\"string\"}],\"outputs\":{\"type\":\"int\"},\"type\":[\"action\",\"event\"]},{\"name\":\"Pause\",\"inputs\":[],\"outputs\":{\"type\":\"int\"},\"type\":[\"action\",\"event\"]},{\"name\":\"Restore\",\"inputs\":[],\"outputs\":{\"type\":\"int\"},\"type\":[\"action\",\"event\"]},{\"name\":\"Destory\",\"inputs\":[],\"outputs\":{\"type\":\"int\"},\"type\":[\"action\",\"event\"]},{\"name\":\"IsSupportToken\",\"inputs\":[{\"name\":\"symbol\",\"type\":\"string\"}],\"outputs\":{\"type\":\"bool\"},\"type\":[]},{\"name\":\"IsStateNormal\",\"inputs\":[],\"outputs\":{\"type\":\"bool\"},\"type\":[]}]")
    if abi == nil {
    	t.Error("create ABI error")
	}
}
