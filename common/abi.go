package common

import "encoding/json"

type Output struct {
	Type string `json:"type"`
}

type Input struct {
	Name string  `json:"name"`
	Type string  `json:"type"`
}

type ABIFuncObject struct {
	Name    string         `json:"name"`
	Inputs  []Input `json:"inputs"`
	Outputs Output         `json:"outputs"`
	Type    []string       `json:"type"`
}

type ABIUtil struct {
	RawABI      string
	FuncObjects []ABIFuncObject
}

func NewABIUtil(rawABI string) *ABIUtil {
	var funcObjects []ABIFuncObject
	err := json.Unmarshal([]byte(rawABI), &funcObjects)
	if err != nil {
		return nil
	}

	return &ABIUtil{rawABI, funcObjects}
}

func (u *ABIUtil) FindPayableMethod()(methodName string, inputs []Input, output *Output) {
	for _, fObj := range u.FuncObjects {
		for _, fObjType := range fObj.Type {
			if fObjType == "payable" {
				return fObj.Name, fObj.Inputs, &fObj.Outputs
			}
		}
	}

	return "", nil, nil
}

func (u *ABIUtil) IsAction(methodName string) bool {
	for _, fObj := range u.FuncObjects {
		if fObj.Name == methodName {
			for _, fObjType := range fObj.Type {
				if fObjType == "action" {
					return true
				}
			}
		}
	}

	return false
}



