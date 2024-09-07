package core

import (
	"davinci/common/log"
	"os"
	"testing"
)

func init() {
	log.AddLogWriter(os.Stdout)
}

func TestElasticSearch_GetNodes(t *testing.T) {

	elastic := &ElasticSearch{
		Url:    "http://192.168.159.135:9200/",
		User:   "kibana",
		Passwd: "123456",
	}
	nodes := elastic.GetNodes()
	log.Output(nodes)
}

func TestElasticSearch_GetIndices(t *testing.T) {
	elastic := &ElasticSearch{
		Url:    "http://192.168.83.129:9200/",
		User:   "elastic",
		Passwd: "123456",
	}
	nodes := elastic.GetIndices()
	log.Output(nodes)
}

func TestElasticSearch_GetMapping(t *testing.T) {
	elastic := &ElasticSearch{
		Url:    "http://192.168.159.135:9200/",
		User:   "elastic",
		Passwd: "123456",
	}
	result := elastic.GetMapping("kibana_sample_data_ecommerce")
	log.Output(result)

}

func TestElasticSearch_GetDocuments(t *testing.T) {
	elastic := &ElasticSearch{
		Url:    "http://192.168.159.135:9200/",
		User:   "elastic",
		Passwd: "123456",
	}
	result := elastic.GetDocuments("kibana_sample_data_ecommerce", 1)
	log.Output(result)

}

func TestElasticSearch_AutoGather(t *testing.T) {
	elastic := &ElasticSearch{
		Url:    "http://192.168.83.129:9200/",
		User:   "elastic",
		Passwd: "12345",
		Check:  true,
	}
	elastic.AutoGather()
}
