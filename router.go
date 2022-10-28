package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type routersConf struct {
	Port    uint32   `json:"port"`
	Routers []router `json:"routers"`
}

type router struct {
	Router  *string `json:"router,omitempty"`
	Disable bool    `json:"disable"`

	To          *string `json:"to,omitempty"`
	ForwardPort *uint32 `json:"forward,omitempty"`
	DistFile    *string `json:"distFile,omitempty"`
}

func getRouters(fp string) *routersConf {
	c := &routersConf{
		Port:    7788,
		Routers: []router{},
	}
	b, err := os.ReadFile(fp)
	if err != nil {
		return c
	}
	err = json.Unmarshal(b, c)
	if err != nil {
		return c
	}

	// for i := 0; i < len(c.Routers); i++ {
	// 	if c.Routers[i].Disable {
	// 		c.Routers = append(c.Routers[:i], c.Routers[i+1:]...)
	// 	}
	// }
	return c

}

type to struct {
	To     string `json:"to"`
	isFile bool   `json:"is_file"`
}

func confToRouters(routers []router) map[string]*to {
	res := make(map[string]*to)
	for _, v := range routers {
		if v.Disable {
			continue
		}
		toStr := ""
		router := func() string {
			if v.Router != nil {
				if v.To != nil {
					toStr = *v.To
				} else {
					if v.DistFile != nil {
						toStr = *v.DistFile
					}
				}
				return *v.Router
			}
			if v.ForwardPort != nil {
				toStr = fmt.Sprintf("%d", *v.ForwardPort)
				return toStr
			}
			return ""
		}()

		if router == "" || toStr == "" {
			continue
		}
		res[router] = &to{
			To:     toStr,
			isFile: v.DistFile != nil,
		}
	}
	return res
}
