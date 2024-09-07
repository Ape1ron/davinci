package core

import (
	"davinci/common"
	"davinci/common/log"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

type ElasticSearch struct {
	Url      string
	User     string
	Passwd   string
	endpoint string
	Check    bool
}

func (e *ElasticSearch) genUrl(path string) string {
	if e.endpoint == "" {
		if strings.HasSuffix(e.Url, "/") {
			e.endpoint = e.Url[:len(e.Url)-1]
		} else {
			e.endpoint = e.Url
		}
		if e.User != "" && e.Passwd != "" {
			basic := fmt.Sprintf("%s:%s", e.User, e.Passwd)
			if strings.HasPrefix(e.endpoint, "https://") {
				e.endpoint = "https://" + basic + "@" + e.endpoint[8:]
			} else {
				e.endpoint = "http://" + basic + "@" + e.endpoint[7:]
			}
		}

	}
	if strings.HasSuffix(path, "/") {
		path = path[:len(path)-1]
	}
	return fmt.Sprintf("%s/%s", e.endpoint, path)

}

func (e *ElasticSearch) ExecuteOnce() {
	log.Warn("ElasticSearch do not support execute mode")
}

func (e *ElasticSearch) Shell() {
	log.Warn("ElasticSearch do not support shell mode")
}

func (e *ElasticSearch) AutoGather() {
	if !e.check() {
		log.Warn("the elasticsearch is not enable")
		return
	}
	log.Output(e.GetNodes())
	log.Output(common.FormatJson(e.GetCount()))
	log.Output(common.FormatJson(e.getUsers()))
	indices := e.GetIndices()
	log.Output(indices)
	for _, index := range common.GetColumnData(indices, 2) {
		log.Output(common.FormatJson(e.GetIndexCount(index)))
		log.Output(common.FormatJson(e.GetMapping(index)))
		log.Output(common.FormatJson(e.getFirst5Docs(index)))
	}
}

func (e *ElasticSearch) SetHost(host string) {
	//c.Host = host
}
func (e *ElasticSearch) SetPort(port int) {
	//c.Port = port
}

func (e *ElasticSearch) SetCmd(cmd string) {

}

func (e *ElasticSearch) check() bool {
	if !e.Check {
		return true
	}
	content := e.getContent("")
	if strings.Contains(strings.ToLower(content), "you know, for search") {
		return true
	} else {
		log.Warn(content)
	}
	return false
}

func (e *ElasticSearch) GetNodes() [][]string {
	log.Info("cluster info")
	path := "_cat/nodes?v"
	return e.getLinesData(path)
}

func (e *ElasticSearch) GetIndices() [][]string {
	log.Info("get all indices")
	path := "_cat/indices?v"
	return e.getLinesData(path)
}

func (e *ElasticSearch) GetMapping(index string) string {
	log.Info("get  index mapping: " + index)
	path := fmt.Sprintf("%s/_mapping", index)
	return e.getContent(path)
}

func (e *ElasticSearch) GetCount() string {
	log.Info("doc count")
	path := "_count"
	return e.getContent(path)
}

func (e *ElasticSearch) GetIndexCount(index string) string {
	log.Info("doc count in: " + index)
	path := fmt.Sprintf("%s/_count", index)
	return e.getContent(path)
}

func (e *ElasticSearch) GetDocuments(index string, size int) string {
	path := fmt.Sprintf("%s/_search?size=%d", index, size)
	return e.getContent(path)
}

func (e *ElasticSearch) getFirst5Docs(index string) string {
	log.Info("get first 5 document in: " + index)
	return e.GetDocuments(index, 5)
}

func (e *ElasticSearch) getUsers() string {
	log.Info("try get users,api only enabled when the auth mode open")
	path := "_xpack/security/user"
	return e.getContent(path)
}

func (e *ElasticSearch) getLinesData(path string) [][]string {
	url := e.genUrl(path)
	method := "GET"
	log.Info(fmt.Sprintf("%s /%s", method, path))
	payload := strings.NewReader(``)
	req, _ := http.NewRequest(method, url, payload)
	var result [][]string
	if rsp := common.Request(req); rsp != nil {
		defer rsp.Body.Close()
		if body, err := ioutil.ReadAll(rsp.Body); err == nil {
			lines := strings.Split(string(body), "\n")
			for _, line := range lines {
				data := common.DelElement(strings.Split(line, " "), "")
				if len(data) != 0 {
					result = append(result, data)
				}
			}
		}
	}
	return result
}

func (e *ElasticSearch) getContent(path string) string {
	url := e.genUrl(path)
	method := "GET"
	log.Info(fmt.Sprintf("%s /%s", method, path))
	payload := strings.NewReader(``)
	req, _ := http.NewRequest(method, url, payload)
	if rsp := common.Request(req); rsp != nil {
		defer rsp.Body.Close()
		if body, err := ioutil.ReadAll(rsp.Body); err == nil {
			return string(body)
		} else {
			log.Warn(err)
		}
	}
	return ""
}
